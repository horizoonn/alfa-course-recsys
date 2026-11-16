import os
import time
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, ConfigDict, Field
from typing import Literal

from search_index import SearchIndex


class SearchRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    query: str = Field(max_length=1000)
    catalog_id: str = Field(pattern=r"^[0-9a-f]{64}$")
    mode: Literal["tfidf", "e5", "rrf"] = "tfidf"


@asynccontextmanager
async def lifespan(app):
    started = time.perf_counter()
    app.state.index = SearchIndex(
        os.environ["SEARCH_INDEX_PATH"],
        enable_semantic=os.environ.get("SEARCH_ENABLE_SEMANTIC", "false").lower()
        == "true",
    )
    app.state.load_ms = (time.perf_counter() - started) * 1000
    yield
    app.state.index = None


app = FastAPI(lifespan=lifespan, docs_url=None, redoc_url=None, openapi_url=None)


@app.get("/health")
def health():
    index = app.state.index
    return {
        "status": "ok",
        "catalog_id": index.manifest["catalog_id"],
        "index_id": index.manifest["generation"],
        "semantic": index.encoder is not None,
        "semantic_error": index.semantic_error,
        "documents": len(index.keys),
        "load_ms": app.state.load_ms,
    }


@app.post("/search")
def search(body: SearchRequest):
    index = app.state.index
    if body.catalog_id != index.manifest["catalog_id"]:
        raise HTTPException(status_code=409, detail="search catalog mismatch")
    return index.search(body.query.strip(), body.mode)


class RelatedRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    material_key: str = Field(max_length=80)
    query: str = Field(default="", max_length=1000)
    catalog_id: str = Field(pattern=r"^[0-9a-f]{64}$")


@app.post("/related")
def related(body: RelatedRequest):
    index = app.state.index
    if body.catalog_id != index.manifest["catalog_id"]:
        raise HTTPException(status_code=409, detail="search catalog mismatch")
    try:
        return index.related(body.material_key, body.query)
    except KeyError:
        raise HTTPException(status_code=404, detail="unknown catalog material")
