"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { useGoAuth } from "@/components/go/GoAuthProvider";
import { GO_ECOMMERCE_URL, getGoAccessToken } from "@/lib/go-auth";

export interface GoCartItem {
  id: string;
  userId: string;
  productId: string;
  quantity: number;
  createdAt: string;
  productName?: string;
  productPrice?: number;
  productImage?: string;
}

interface GoCartContextType {
  items: GoCartItem[];
  count: number;
  refresh: () => Promise<void>;
}

const GoCartContext = createContext<GoCartContextType>({
  items: [],
  count: 0,
  refresh: async () => {},
});

export function useGoCart() {
  return useContext(GoCartContext);
}

export function GoCartProvider({ children }: { children: React.ReactNode }) {
  const { isLoggedIn } = useGoAuth();
  const [items, setItems] = useState<GoCartItem[]>([]);

  const refresh = useCallback(async () => {
    const token = getGoAccessToken();
    if (!token) {
      setItems([]);
      return;
    }
    try {
      const res = await fetch(`${GO_ECOMMERCE_URL}/cart`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) return;
      const data = await res.json();
      setItems(data.items ?? []);
    } catch {
      /* swallow — badge stays stale on network failure */
    }
  }, []);

  useEffect(() => {
    if (isLoggedIn) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      refresh();
    } else {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setItems([]);
    }
  }, [isLoggedIn, refresh]);

  const count = items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <GoCartContext.Provider value={{ items, count, refresh }}>
      {children}
    </GoCartContext.Provider>
  );
}
