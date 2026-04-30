/**
 * @vitest-environment jsdom
 */
import { describe, it, expect } from "vitest";
import { evaluateRetry } from "./retry";

describe("evaluateRetry", () => {
  it("returns true when failureCount < 3 and error has retry", () => {
    expect(evaluateRetry(0, { status: 500, retry: true })).toBe(true);
    expect(evaluateRetry(1, { status: 401, retry: true })).toBe(true);
    expect(evaluateRetry(2, { status: 502, retry: true })).toBe(true);
  });

  it("returns false when failureCount >= 3", () => {
    expect(evaluateRetry(3, { status: 500, retry: true })).toBe(false);
    expect(evaluateRetry(4, { status: 500, retry: true })).toBe(false);
  });

  it("returns false when error has no retry", () => {
    expect(evaluateRetry(0, { status: 400, retry: false })).toBe(false);
    expect(evaluateRetry(1, { status: 404 })).toBe(false);
  });

  it("does not retry for 4xx errors other than 401 even if retry is true", () => {
    expect(evaluateRetry(0, { status: 400, retry: true })).toBe(false);
    expect(evaluateRetry(0, { status: 404, retry: true })).toBe(false);
    expect(evaluateRetry(0, { status: 422, retry: true })).toBe(false);
  });
});
