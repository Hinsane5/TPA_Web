from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from transformers import pipeline

app = FastAPI()
MODEL_NAME = "Maungvee/baption-summarizer" 

print("Loading AI Model... this might take a minute...")
try:
    summarizer = pipeline("summarization", model=MODEL_NAME, device=-1)
    print("AI Model Loaded Successfully!")
except Exception as e:
    print(f"Error loading model: {e}")
    summarizer = None

class CaptionRequest(BaseModel):
    text: str

@app.post("/summarize")
async def summarize_caption(request: CaptionRequest):
    if not summarizer:
        raise HTTPException(status_code=503, detail="AI Model is not ready yet.")
    
    try:
        summary = summarizer(request.text, max_length=50, min_length=10, do_sample=False)
        return {"summary": summary[0]['summary_text']}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
def health_check():
    return {"status": "ok", "model_loaded": summarizer is not None}