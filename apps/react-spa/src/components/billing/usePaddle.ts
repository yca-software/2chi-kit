import { useEffect } from "react";
import type { CheckoutEventsData } from "@paddle/paddle-js";
import {
  usePaddleContext,
  type CheckoutCompletedContext,
} from "./PaddleProvider";

export type { CheckoutCompletedContext };

export interface UsePaddleOptions {
  /** Called when checkout completes (checkout.completed). Set once when the hook runs; keep this stable or the handler will be updated. */
  onCheckoutCompleted?: (
    data: CheckoutEventsData,
    context: CheckoutCompletedContext | null,
  ) => void;
}

/**
 * Use Paddle checkout. Must be used within PaddleProvider.
 * Pass onCheckoutCompleted to run logic when the user completes checkout (e.g. process-transaction API).
 */
export function usePaddle(options: UsePaddleOptions = {}) {
  const ctx = usePaddleContext();
  const { onCheckoutCompleted } = options;

  useEffect(() => {
    if (onCheckoutCompleted) {
      ctx.setCheckoutCompletedHandler(onCheckoutCompleted);
    }
  }, [onCheckoutCompleted, ctx]);

  return {
    ready: ctx.ready,
    error: ctx.error,
    openCheckoutWithTransactionId: ctx.openCheckoutWithTransactionId,
    openCheckoutWithItems: ctx.openCheckoutWithItems,
  };
}
