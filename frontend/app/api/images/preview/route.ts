import { forwardImageRequest } from "../forward-image-request";
export async function POST(request: Request) {
 return forwardImageRequest(request, "/images/preview");
}
