import test from "node:test";
import assert from "node:assert/strict";
import { acquirePreview, nativeDisplayBlob } from "../app/components/image-upload/preview-cache.ts";

test("native AVIF MIME correction preserves original", async () => {
 const source = new Blob([new Uint8Array([0,0,0,24]), "ftypavif", new Uint8Array(12)], {type:"application/octet-stream"});
 const display = await nativeDisplayBlob(source);
 assert.equal(display.type,"image/avif");
 assert.equal(source.type,"application/octet-stream");
 assert.deepEqual(await display.arrayBuffer(), await source.arrayBuffer());
});
test("result source, concurrent sharing, reuse and explicit retry", async () => {
 const originalFetch = globalThis.fetch;
 const source = new Blob(["processed-result"], {type:"image/tiff"});
 let calls = 0;
 globalThis.fetch = async (url, options) => {
  calls++;
  assert.equal(url,"/api/images/preview");
  assert.equal(await options.body.get("image").text(),"processed-result");
  return new Response("preview",{headers:{"Content-Type":"image/png"}});
 };
 try {
  const a = acquirePreview(source), b = acquirePreview(source);
  assert.equal(await a.promise,await b.promise); a.release(); b.release();
  const c = acquirePreview(source); await c.promise; c.release(); assert.equal(calls,1);
  globalThis.fetch = async () => { calls++; return Response.json({error:"unsupported"},{status:422}); };
  const bad = new Blob(["bad"]);
  const d = acquirePreview(bad); await assert.rejects(d.promise,/unsupported/); d.release();
  const e = acquirePreview(bad); await assert.rejects(e.promise); e.release(); assert.equal(calls,2);
  const f = acquirePreview(bad,true); await assert.rejects(f.promise); f.release(); assert.equal(calls,3);
 } finally { globalThis.fetch = originalFetch; }
});
test("last consumer aborts obsolete request and replacement can start", async () => {
 const previous = globalThis.fetch;
 let signals = [];
 globalThis.fetch = async (_,options) => {
  signals.push(options.signal);
  return new Promise((resolve,reject) => options.signal.addEventListener("abort", () => reject(new DOMException("Aborted","AbortError"))));
 };
 try {
  const source = new Blob(["source"]);
  const a = acquirePreview(source), b = acquirePreview(source);
  const rejected = assert.rejects(a.promise);
  a.release(); assert.equal(signals[0].aborted,false);
  b.release(); assert.equal(signals[0].aborted,true); await rejected;
  const c = acquirePreview(source); const canceled = assert.rejects(c.promise); c.release(); await canceled;
  assert.equal(signals.length,2);
 } finally { globalThis.fetch = previous; }
});
