#!/usr/bin/env python3
"""Sync the Grafana system-overview dashboard JSON into the K8s ConfigMap.

The dashboard source of truth lives at
``monitoring/grafana/dashboards/system-overview.json``. The same JSON is
embedded verbatim into ``k8s/monitoring/configmaps/grafana-dashboards.yml``
under ``data."system-overview.json"`` with 4-space indentation.

CI fails if these drift. This script regenerates the ConfigMap from the
source JSON, or with ``--check`` verifies they're in sync without writing.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

DASHBOARD = Path("monitoring/grafana/dashboards/system-overview.json")
CONFIGMAP = Path("k8s/monitoring/configmaps/grafana-dashboards.yml")

HEADER = (
    "apiVersion: v1\n"
    "kind: ConfigMap\n"
    "metadata:\n"
    "  name: grafana-dashboards\n"
    "  namespace: monitoring\n"
    "data:\n"
    "  system-overview.json: |\n"
)


def regenerate() -> None:
    src = DASHBOARD.read_text()
    json.loads(src)  # validate
    indented = "\n".join(
        ("    " + line) if line else line for line in src.splitlines()
    )
    CONFIGMAP.write_text(HEADER + indented + "\n")
    print("regenerated")


def check() -> int:
    import yaml  # only needed for the check path

    src = json.loads(DASHBOARD.read_text())
    cm = yaml.safe_load(CONFIGMAP.read_text())
    embedded = json.loads(cm["data"]["system-overview.json"])
    if src != embedded:
        print(
            "grafana dashboard drift — run: make grafana-sync",
            file=sys.stderr,
        )
        return 1
    print("grafana dashboards in sync")
    return 0


def main() -> int:
    if len(sys.argv) > 1 and sys.argv[1] == "--check":
        return check()
    regenerate()
    return 0


if __name__ == "__main__":
    sys.exit(main())
