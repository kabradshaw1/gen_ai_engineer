"use client";

import { usePathname } from "next/navigation";
import { GoCartIcon } from "@/components/go/GoCartIcon";
import { GoUserDropdown } from "@/components/go/GoUserDropdown";
import { useGoStore } from "@/components/go/GoStoreProvider";

export function GoSubHeader() {
  const pathname = usePathname();
  const onStore = pathname === "/go/ecommerce";
  const { categories, activeCategory, setActiveCategory } = useGoStore();

  return (
    <div className="border-b border-foreground/10 bg-background">
      <div className="mx-auto flex h-12 max-w-5xl items-center justify-between px-6">
        <div className="flex items-center gap-4">
          {onStore && <h1 className="text-lg font-semibold">Store</h1>}
        </div>
        <div className="flex items-center gap-4">
          <GoCartIcon />
          <GoUserDropdown />
        </div>
      </div>
      {onStore && (
        <div className="border-t border-foreground/5">
          <div className="mx-auto flex max-w-5xl flex-wrap items-center gap-2 px-6 py-2">
            <button
              onClick={() => setActiveCategory(null)}
              className={`rounded-full px-3 py-1 text-sm transition-colors ${
                activeCategory === null
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-muted-foreground hover:text-foreground"
              }`}
            >
              All
            </button>
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setActiveCategory(cat)}
                className={`rounded-full px-3 py-1 text-sm transition-colors ${
                  activeCategory === cat
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground hover:text-foreground"
                }`}
              >
                {cat}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
