import test from "node:test";
import assert from "node:assert/strict";
import { conversionSizeChange } from "../app/components/image-convert/size-change.ts";
test("measured conversion size outcomes", () => {
 assert.deepEqual(conversionSizeChange(100,50), {label:"Reduction", value:"50.0%"});
 assert.deepEqual(conversionSizeChange(100,100), {label:"Size change", value:"No size change"});
 assert.deepEqual(conversionSizeChange(100,150), {label:"Increase", value:"50.0%"});
 assert.deepEqual(conversionSizeChange(10000,10001), {label:"Increase", value:"<0.1%"});
 assert.deepEqual(conversionSizeChange(10000,9999), {label:"Reduction", value:"<0.1%"});
});
