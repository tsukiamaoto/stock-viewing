import os
import json
import google.generativeai as genai
from dotenv import load_dotenv

load_dotenv()

API_KEY = os.getenv("GEMINI_API_KEY")
if API_KEY:
    genai.configure(api_key=API_KEY)

# 系統 Prompt 告訴模型要怎麼分類新聞
SYSTEM_PROMPT = """
你是一個專業的金融新聞分類 AI。
使用者會提供一組新聞列表（包含標題、摘要），你需要將這些新聞分類到以下五個特定主題中：
1. "trump": 川普的相關發言或政策
2. "hormuz_iran": 荷姆茲海峽、伊朗、中東地緣政治相關新聞
3. "ai": AI 相關技術、半導體基礎設施、人工智慧新聞
4. "finance": 一般財經大盤、降息、外資報告等新聞
5. "other": 無法歸類到上述四類的新聞（前端可能不顯示）

請以 JSON 格式輸出，格式必須為：
{
  "categories": {
    "trump": [ {"title": "...", "snippet": "..."} ],
    "hormuz_iran": [ ... ],
    "ai": [ ... ],
    "finance": [ ... ]
  }
}
請只輸出合法的 JSON 字串，不要包含 ```json 標籤或其他對話文字。
"""

def categorize_news_with_llm(news_list: list) -> dict:
    """使用 LLM 將新聞分類"""
    if not API_KEY:
        # 如果沒有設定 API KEY，為了順利展示 MVP，提供本地模擬的分類結果（或寫個簡單規則分類）
        print("Warning: GEMINI_API_KEY not found. Using mock fallback classifier.")
        return mock_fallback_classifier(news_list)
        
    try:
        model = genai.GenerativeModel('gemini-1.5-flash', system_instruction=SYSTEM_PROMPT)
        
        # 將新聞清單轉成文本傳給 LLM
        prompt_text = "以下是待分類的新聞列表：\n"
        for i, n in enumerate(news_list):
            prompt_text += f"[{i+1}] 標題: {n.get('title')}\n摘要: {n.get('snippet')}\n來源: {n.get('source')}\n\n"
            
        response = model.generate_content(prompt_text)
        
        # 解析 JSON
        result_text = response.text.strip()
        if result_text.startswith("```json"):
            result_text = result_text[7:-3].strip()
        elif result_text.startswith("```"):
            result_text = result_text[3:-3].strip()
            
        return json.loads(result_text)
        
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
