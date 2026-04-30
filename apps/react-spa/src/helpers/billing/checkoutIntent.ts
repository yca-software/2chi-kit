export interface CheckoutIntentPayload {
  checkoutIntent: string;
  priceId: string | null;
}

const CHECKOUT_INTENT_KEY = "checkout_intent_payload";

export function captureCheckoutIntentFromSearchParams(search: string): void {
  if (!search) return;
  const params = new URLSearchParams(search);
  const checkoutIntent = params.get("checkoutIntent") ?? params.get("plan");
  if (!checkoutIntent) return;
  const payload: CheckoutIntentPayload = {
    checkoutIntent,
    priceId: params.get("priceId"),
  };
  sessionStorage.setItem(CHECKOUT_INTENT_KEY, JSON.stringify(payload));
}

export function consumeCheckoutIntent(): CheckoutIntentPayload | null {
  const raw = sessionStorage.getItem(CHECKOUT_INTENT_KEY);
  if (!raw) return null;
  sessionStorage.removeItem(CHECKOUT_INTENT_KEY);
  try {
    const parsed = JSON.parse(raw) as Partial<CheckoutIntentPayload>;
    if (!parsed.checkoutIntent) return null;
    return {
      checkoutIntent: parsed.checkoutIntent,
      priceId: parsed.priceId ?? null,
    };
  } catch {
    return null;
  }
}
