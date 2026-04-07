"use client";

import Link from "next/link";
import { ShoppingCart } from "lucide-react";
import { useGoAuth } from "@/components/go/GoAuthProvider";
import { useGoCart } from "@/components/go/GoCartProvider";

export function GoCartIcon() {
  const { isLoggedIn } = useGoAuth();
  const { count } = useGoCart();

  if (!isLoggedIn) return null;

  return (
    <Link
      href="/go/ecommerce/cart"
      className="relative text-muted-foreground hover:text-foreground transition-colors"
      aria-label="Cart"
    >
      <ShoppingCart className="size-5" />
      {count > 0 && (
        <span className="absolute -right-2 -top-2 flex size-4 items-center justify-center rounded-full bg-foreground text-[10px] font-semibold text-background">
          {count > 99 ? "99+" : count}
        </span>
      )}
    </Link>
  );
}
