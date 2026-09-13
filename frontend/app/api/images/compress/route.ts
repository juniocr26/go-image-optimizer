export const runtime = "nodejs";

const MAX_UPLOAD_BYTES = 50 * 1024 * 1024;

export async function POST(request: Request) {
  const backendUrl = process.env.BACKEND_URL ?? "http://localhost:8080";

  let formData: FormData;

  try {
    formData = await request.formData();
  } catch {
    return Response.json({ error: "Invalid multipart form." }, { status: 400 });
  }

  const image = formData.get("image");
  if (image instanceof File && image.size > MAX_UPLOAD_BYTES) {
    return Response.json(
      { error: "The selected image is larger than 50 MiB." },
      { status: 413 },
    );
  }

  try {
    const backendResponse = await fetch(`${backendUrl}/images/compress`, {
      method: "POST",
      body: formData,
    });

    const contentType =
      backendResponse.headers.get("Content-Type") ?? "application/octet-stream";

    if (!backendResponse.ok) {
      const errorBody = await backendResponse.text();

      return new Response(errorBody, {
        status: backendResponse.status,
        headers: {
          "Content-Type": contentType,
        },
      });
    }

    const imageBytes = await backendResponse.arrayBuffer();
    const headers = new Headers({
      "Content-Length": String(imageBytes.byteLength),
      "Content-Type": contentType,
    });

    const contentDisposition = backendResponse.headers.get("Content-Disposition");
    if (contentDisposition) {
      headers.set("Content-Disposition", contentDisposition);
    }

    return new Response(imageBytes, {
      status: backendResponse.status,
      headers,
    });
  } catch {
    return Response.json({ error: "Backend is unavailable." }, { status: 502 });
  }
}
