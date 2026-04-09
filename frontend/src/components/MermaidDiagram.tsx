"use client";

import { useEffect, useId, useRef } from "react";
import mermaid from "mermaid";
import DOMPurify from "dompurify";

mermaid.initialize({
  startOnLoad: false,
  theme: "dark",
  themeVariables: {
    darkMode: true,
  },
});

let counter = 0;

export function MermaidDiagram({ chart }: { chart: string }) {
  const ref = useRef<HTMLDivElement>(null);
  const reactId = useId();

  useEffect(() => {
    if (!ref.current) return;

    // useId() returns strings like ":r0:" which contain colons that are
    // invalid in mermaid element IDs. Combine with a counter for uniqueness.
    const id = `mermaid-${reactId.replace(/:/g, "")}-${counter++}`;
    mermaid.render(id, chart).then(({ svg }) => {
      if (ref.current) {
        const clean = DOMPurify.sanitize(svg, {
          USE_PROFILES: { svg: true, svgFilters: true },
        });
        ref.current.innerHTML = clean;
      }
    });
  }, [chart, reactId]);

  return <div ref={ref} className="flex justify-center overflow-x-auto" />;
}
