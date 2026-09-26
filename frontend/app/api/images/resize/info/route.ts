import { forwardImageRequest } from "../../forward-image-request";
export const runtime = "nodejs";
export async function POST(request: Request) {
  return forwardImageRequest(request, "/images/resize/info");
}
