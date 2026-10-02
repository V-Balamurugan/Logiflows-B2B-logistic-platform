import logging
import os
import time
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from app.api import health, predict

logging.basicConfig(
    level=logging.INFO,
    format='{"time": "%(asctime)s", "level": "%(levelname)s", "name": "%(name)s", "message": "%(message)s"}'
)
logger = logging.getLogger("logiflows.ai")

app = FastAPI(
    title="LogiFlows AI Advisory Service",
    description="Advisory AI service for delivery delay risk prediction and route intelligence",
    version="0.0.1-foundation"
)

@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time)
    return response

app.include_router(health.router, tags=["Health"])
app.include_router(predict.router, prefix="/api/v1/predict", tags=["Prediction"])

@app.get("/")
def root():
    return {
        "service": "LogiFlows AI Advisory Service",
        "role": "Advisory Delay Intelligence & Feature Evaluation",
        "authority": "Non-Authoritative (Advisory Only)",
        "status": "operational"
    }
