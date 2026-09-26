import assert from "node:assert/strict";
import test from "node:test";
import {
  targetDimensions,
  dimensionError,
} from "../app/components/image-resize/resize-options.ts";

const info = { width: 899, height: 1599, frameCount: 1 };
const defaults = {
  mode: "pixels",
  width: "899",
  height: "1599",
  axis: "width",
  keepAspectRatio: true,
  reduction: 50,
};

for (const [name, overrides, expected] of [
  ["unchanged", {}, { width: 899, height: 1599 }],
  ["enlarge from width", { width: "1798", height: "1" }, { width: 1798, height: 3198 }],
  ["enlarge from height", { width: "1", height: "3198", axis: "height" }, { width: 1798, height: 3198 }],
  ["independent enlargement and reduction", { width: "1798", height: "200", keepAspectRatio: false }, { width: 1798, height: 200 }],
  ["25% smaller", { mode: "percentage", reduction: 25 }, { width: 674, height: 1199 }],
  ["50% smaller", { mode: "percentage", reduction: 50 }, { width: 450, height: 800 }],
  ["75% smaller", { mode: "percentage", reduction: 75 }, { width: 225, height: 400 }],
]) {
  test(name, () => {
    const target = targetDimensions(info, { ...defaults, ...overrides });
    assert.deepEqual(target, expected);
    assert.equal(dimensionError(info, target), null);
  });
}

test("invalid dimensions and resource budgets prevent submission", () => {
  for (const width of ["", "0", "-1", "1.5", "NaN", "32000001"]) {
    assert.equal(targetDimensions(info, { ...defaults, width }), null);
  }
  const target = targetDimensions(info, {
    ...defaults, width: "10000", height: "10000", keepAspectRatio: false,
  });
  assert.match(dimensionError(info, target), /32 million/);
  assert.match(dimensionError({ ...info, frameCount: 100 }, { width: 1000, height: 1000 }), /frame pixel/);
  assert.deepEqual(targetDimensions({ width: 1, height: 1, frameCount: 1 }, {
    ...defaults, mode: "percentage", reduction: 75,
  }), { width: 1, height: 1 });
});
