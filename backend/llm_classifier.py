import json
import google.generativeai as genai
import random
import time
from config import GEMINI_API_KEY as API_KEY, GEMINI_MODELS as MODELS, SYSTEM_PROMPT, ENHANCE_PROMPT
MODEL_INDEX = 0

# 記錄已知限額用盡的模型，暫時跳過
_EXHAUSTED_MODELS: set = set()

if API_KEY:
    genai.configure(api_key=API_KEY)


def _build_model_and_prompt(model_name: str, system_prompt: str, user_text: str):
    """
    為了避免因各模型（如 Gemma 3）對 system_instruction 的支援不一致導致 400 錯誤，
    這邊一律將指令結合並傳遞給 user message。
    """
    model = genai.GenerativeModel(model_name)
    combined_prompt = f"{system_prompt.strip()}\n\n{user_text}"
    return model, combined_prompt


def _extract_json(text: str) -> dict:
    """從一段可能夾雜垃圾文字的 LLM 輸出中擷取並解析第一個 JSON 物件"""
    start_idx = text.find('{')
    if start_idx == -1:
        raise ValueError("No JSON object found: '{' is missing.")
    
    depth = 0
    for i in range(start_idx, len(text)):
        if text[i] == '{':
            depth += 1
        elif text[i] == '}':
            depth -= 1
            if depth == 0:
                json_str = text[start_idx:i+1]
                return json.loads(json_str, strict=False)
                
    raise ValueError("No complete JSON object found: unmatched '{'.")


def categorize_news_with_llm(news_list: list) -> dict:
    """使用 LLM 將新聞分類"""
    if not API_KEY:
        print("Warning: GEMINI_API_KEY not found. Using mock fallback classifier.")
        return mock_fallback_classifier(news_list)

    try:
        current_model = MODELS[MODEL_INDEX % len(MODELS)]
        prompt_text = "以下是待分類的新聞列表：\n"
        for i, n in enumerate(news_list):
            prompt_text += f"[{i+1}] 標題: {n.get('title')}\n摘要: {n.get('snippet')}\n來源: {n.get('source')}\n\n"

        model, full_prompt = _build_model_and_prompt(current_model, SYSTEM_PROMPT, prompt_text)
        response = model.generate_content(full_prompt)

        result_text = response.text.strip()
        
        # 尋找並抽取第一個完整的 JSON 區塊
        return _extract_json(result_text)

    except Exception as e:
        print(f"LLM Classification Error: {e}")
        return mock_fallback_classifier(news_list)


def mock_fallback_classifier(news_list: list) -> dict:
    """簡單的關鍵字分類器，當沒有 API_KEY 時作為 fallback"""
    result = {
        "categories": {
            "trump": [],
            "hormuz_iran": [],
            "ai": [],
            "finance": []
        }
    }
    
    for n in news_list:
        text = str(n.get("title", "")) + " " + str(n.get("snippet", ""))
        
        if "川普" in text or "Trump" in text:
            result["categories"]["trump"].append(n)
        elif "伊朗" in text or "荷姆茲" in text or "中東" in text:
            result["categories"]["hormuz_iran"].append(n)
        elif "AI" in text or "半導體" in text or "晶片" in text or "人工智慧" in text:
            result["categories"]["ai"].append(n)
        else:
            result["categories"]["finance"].append(n)
            
    return result



def _is_rate_limit_error(e: Exception) -> bool:
    """判斷是否是限額相關的錯誤（429、Resource exhausted 等）"""
    msg = str(e).lower()
    return any(kw in msg for kw in ["429", "resource_exhausted", "quota", "rate limit", "too many requests"])


def enhance_news_with_llm(title: str, snippet: str) -> dict:
    """
    單篇新聞交給 LLM 做分類與翻譯。
    若收到限額錯誤 (429)，立即換下一個模型重試同一篇文章。
    若是其他錯誤（網路、格式等），也換下一個模型。
    """
    if not API_KEY:
        print("Warning: GEMINI_API_KEY not found. Using fallback.")
        return {
            "category": "other",
            "translated_title": title,
            "translated_snippet": snippet
        }

    global MODEL_INDEX, _EXHAUSTED_MODELS

    # 取得可用模型列表（排除暫時記錄為限額耗盡的）
    available = [m for m in MODELS if m not in _EXHAUSTED_MODELS]
    if not available:
        print("[LLM] All models exhausted this session. Clearing and retrying full list.")
        _EXHAUSTED_MODELS.clear()
        available = MODELS[:]

    tried = set()
    for _ in range(len(available)):
        model_name = available[MODEL_INDEX % len(available)]
        MODEL_INDEX += 1

        if model_name in tried:
            continue
        tried.add(model_name)

        try:
            user_text = f"新聞標題: {title}\n新聞內文: {snippet}\n"
            model, full_prompt = _build_model_and_prompt(model_name, ENHANCE_PROMPT, user_text)

            response = model.generate_content(full_prompt)
            result_text = response.text.strip()
            
            # 使用更穩定的跨號追蹤避免 extra data 錯誤
            data = _extract_json(result_text)
            return data

        except Exception as e:
            if _is_rate_limit_error(e):
                print(f"[LLM] {model_name} 額度用盡 (429)，暫時移除並換下一個模型...")
                _EXHAUSTED_MODELS.add(model_name)
            else:
                print(f"[LLM] {model_name} 發生錯誤: {e}，嘗試下一個模型...")
            time.sleep(0.5)

    print("[LLM] 所有可用模型均失敗，回傳原文。")
    return {
        "category": "other",
        "translated_title": title,
        "translated_snippet": snippet
    }
