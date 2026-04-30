import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import { useQueryClient } from "@tanstack/react-query";
import {
  initializePaddle,
  type Paddle,
  type PaddleEventData,
  type CheckoutEventsData,
  CheckoutEventNames,
} from "@paddle/paddle-js";
import { PADDLE_CLIENT_TOKEN, PADDLE_ENVIRONMENT } from "@/constants";
import {
  ORGANIZATION_QUERY_KEYS,
  USER_QUERY_KEYS,
} from "@/constants/queryKeys";
import { getPreferredLanguage } from "@/helpers/language";

/** Context passed when checkout was opened with openCheckoutWithTransactionId(transactionId, context). */
export interface CheckoutCompletedContext {
  organizationId: string;
  priceId: string;
}

type CheckoutCompletedHandler = (
  data: CheckoutEventsData,
  context: CheckoutCompletedContext | null,
) => void;

interface PaddleContextValue {
  ready: boolean;
  error: string | null;
  openCheckoutWithTransactionId: (
    transactionId: string,
    context?: CheckoutCompletedContext | null,
  ) => void;
  openCheckoutWithItems: (priceId: string, customerEmail?: string) => void;
  /** Set the handler that runs when checkout completes. Called by usePaddle when options.onCheckoutCompleted is provided. */
  setCheckoutCompletedHandler: (
    handler: CheckoutCompletedHandler | null,
  ) => void;
}

const PaddleContext = createContext<PaddleContextValue | null>(null);

export function PaddleProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [paddle, setPaddle] = useState<Paddle | undefined>(undefined);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const locale = getPreferredLanguage();

  const completionRef = useRef<{
    handler: CheckoutCompletedHandler | null;
    context: CheckoutCompletedContext | null;
  }>({ handler: null, context: null });

  useEffect(() => {
    if (!PADDLE_CLIENT_TOKEN) {
      setError("Paddle is not configured");
      return;
    }
    let cancelled = false;
    initializePaddle({
      token: PADDLE_CLIENT_TOKEN,
      environment: PADDLE_ENVIRONMENT,
      eventCallback: (event: PaddleEventData) => {
        if (event.name !== CheckoutEventNames.CHECKOUT_COMPLETED) return;

        queryClient.invalidateQueries({
          queryKey: [ORGANIZATION_QUERY_KEYS.ALL],
        });
        queryClient.invalidateQueries({
          queryKey: [ORGANIZATION_QUERY_KEYS.DETAIL],
        });
        queryClient.invalidateQueries({
          queryKey: [USER_QUERY_KEYS.CURRENT],
        });

        const { handler, context } = completionRef.current;
        completionRef.current = { handler: null, context: null };
        if (event.data && handler) {
          handler(event.data, context);
        }
      },
    })
      .then((instance) => {
        if (cancelled) return;
        if (instance) {
          setPaddle(instance);
          setReady(true);
        } else {
          setError("Paddle failed to load");
        }
      })
      .catch((err) => {
        if (!cancelled) setError(err?.message ?? "Failed to load Paddle");
      });
    return () => {
      cancelled = true;
    };
  }, [queryClient]);

  const setCheckoutCompletedHandler = useCallback(
    (handler: CheckoutCompletedHandler | null) => {
      completionRef.current.handler = handler;
    },
    [],
  );

  const openCheckoutWithTransactionId = useCallback(
    (transactionId: string, context?: CheckoutCompletedContext | null) => {
      if (!paddle) return;
      completionRef.current.context = context ?? null;
      paddle.Checkout.open({
        transactionId,
        settings: { locale },
      });
    },
    [paddle, locale],
  );

  const openCheckoutWithItems = useCallback(
    (priceId: string, customerEmail?: string) => {
      if (!paddle) return;
      paddle.Checkout.open({
        items: [{ priceId, quantity: 1 }],
        ...(customerEmail && { customer: { email: customerEmail } }),
        settings: { locale },
      });
    },
    [paddle, locale],
  );

  const value = useMemo<PaddleContextValue>(
    () => ({
      ready,
      error,
      openCheckoutWithTransactionId,
      openCheckoutWithItems,
      setCheckoutCompletedHandler,
    }),
    [
      ready,
      error,
      openCheckoutWithTransactionId,
      openCheckoutWithItems,
      setCheckoutCompletedHandler,
    ],
  );

  return (
    <PaddleContext.Provider value={value}>{children}</PaddleContext.Provider>
  );
}

export function usePaddleContext(): PaddleContextValue {
  const ctx = useContext(PaddleContext);
  if (!ctx) {
    throw new Error("usePaddleContext must be used within PaddleProvider");
  }
  return ctx;
}
