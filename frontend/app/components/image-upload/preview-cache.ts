// Cache display-only bytes by source identity; originals/results stay untouched.
type Entry = { controller: AbortController; users: number; promise: Promise<Blob>; blob?: Blob; failed?: boolean };
const cache = new WeakMap<Blob, Entry>();

export function acquirePreview(source: Blob, retry = false) {
  let entry = cache.get(source);
  if (retry && entry?.failed) { cache.delete(source); entry = undefined; }
  if (!entry) {
    const controller = new AbortController();
    const form = new FormData();
    form.append("image", source, source instanceof File ? source.name : "result");
    const current: Entry = { controller, users: 0, promise: Promise.resolve(source) };
    current.promise = fetch("/api/images/preview", { method: "POST", body: form, signal: controller.signal })
      .then(async response => {
        if (!response.ok) {
          const data = await response.json();
          throw new Error(data.error || "Preview could not be generated.");
        }
        const blob = await response.blob();
        if (!blob.size || !["image/png", "image/jpeg"].includes(blob.type)) throw new Error("Invalid preview response.");
        current.blob = blob;
        return blob;
      }).catch(error => { current.failed = true; throw error; });
    entry = current;
    cache.set(source, entry);
  }
  entry.users++;
  const acquired = entry;
  let released = false;
  return { promise: acquired.promise, release() {
    if (released) return;
    released = true;
    acquired.users--;
    if (!acquired.users && !acquired.blob && !acquired.failed) {
      acquired.controller.abort();
      if (cache.get(source) === acquired) cache.delete(source);
    }
  } };
}

// Correct mislabeled bytes for native decoding without altering file metadata.
export async function nativeDisplayBlob(source: Blob): Promise<Blob> {
  const bytes = new Uint8Array(await source.slice(0, 64).arrayBuffer());
  const text = new TextDecoder("latin1").decode(bytes);
  let type = source.type;
  if (bytes[0] === 0xff && bytes[1] === 0xd8) type = "image/jpeg";
  else if (bytes[0] === 137 && text.slice(1,4) === "PNG") type = "image/png";
  else if (text.startsWith("GIF8")) type = "image/gif";
  else if (text.startsWith("RIFF") && text.slice(8,12) === "WEBP") type = "image/webp";
  else if (text.slice(4,8) === "ftyp" && /avif|avis/.test(text.slice(8))) type = "image/avif";
  return type === source.type ? source : source.slice(0, source.size, type);
}
