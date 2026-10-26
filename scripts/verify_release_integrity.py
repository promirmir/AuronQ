#!/usr/bin/env python3
"""Read-only monitoring of exact official GitHub Release tags and asset bytes.

Unlike SHA256SUMS uploaded with a release, the expected digests live in an
independently reviewable Git commit. This detects changes, but cannot stop
an administrator from changing GitHub settings or published artifacts.
No wallet or blockchain data is read or modified.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import sys
from urllib.parse import quote
from urllib.request import Request, urlopen


class IntegrityError(Exception):
    """A published release no longer matches its pinned public record."""


def manifest_load(path):
    data = json.loads(Path(path).read_text(encoding="utf-8"))
    if data.get("schema") != 1 or data.get("repository") != "promirmir/AuronQ":
        raise IntegrityError("unexpected baseline schema/repository")
    releases = data.get("releases")
    if not isinstance(releases, list) or len(releases) != 6:
        raise IntegrityError("baseline must cover six pinned historical and current official releases")
    seen = set()
    for item in releases:
        tag = item.get("tag")
        if not isinstance(tag, str) or not re.fullmatch(r"[A-Za-z0-9._-]+", tag) or tag in seen:
            raise IntegrityError("duplicate/invalid tag in baseline")
        seen.add(tag)
        if not re.fullmatch(r"[0-9a-f]{40}", item.get("commit", "")):
            raise IntegrityError("bad expected source commit")
        if type(item.get("release_id")) is not int or item["release_id"] <= 0:
            raise IntegrityError("bad GitHub Release identity")
        assets = item.get("assets")
        if not isinstance(assets, dict) or not assets:
            raise IntegrityError("empty asset baseline")
        for name, meta in assets.items():
            if not name or name in {".", ".."} or "/" in name or "\\" in name:
                raise IntegrityError("unsafe asset name")
            if not re.fullmatch(r"[0-9a-f]{64}", meta.get("sha256", "")):
                raise IntegrityError("invalid pinned SHA-256")
            if type(meta.get("size")) is not int or not 0 < meta["size"] < 1_000_000_000:
                raise IntegrityError("invalid asset size")
    return data


def github_json(repo, path):
    url = f"https://api.github.com/repos/{repo}/{path}"
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "AuronQ-release-integrity-watch/1",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    token = os.environ.get("GITHUB_TOKEN", "")
    if token:
        headers["Authorization"] = "Bearer " + token
    with urlopen(Request(url, headers=headers), timeout=30) as response:
        return json.load(response)


def verify_metadata(expected, ref, release):
    """Fail closed on shifted/deleted tags or changed/recreated release assets."""
    tag = expected["tag"]
    if ref.get("ref") != f"refs/tags/{tag}":
        raise IntegrityError(f"{tag}: unexpected tag reference")
    obj = ref.get("object", {})
    if obj.get("type") != "commit" or obj.get("sha") != expected["commit"]:
        raise IntegrityError(f"{tag}: tag moved away from pinned source commit")
    if release.get("tag_name") != tag or release.get("id") != expected["release_id"]:
        raise IntegrityError(f"{tag}: missing, renamed or recreated GitHub Release")
    if release.get("draft") or not release.get("published_at"):
        raise IntegrityError(f"{tag}: official published release is no longer public")
    actual_assets = release.get("assets", [])
    names = [a.get("name") for a in actual_assets]
    if len(names) != len(set(names)) or set(names) != set(expected["assets"]):
        raise IntegrityError(f"{tag}: release asset inventory changed")
    for asset in actual_assets:
        name = asset["name"]
        pin = expected["assets"][name]
        if asset.get("digest") != "sha256:" + pin["sha256"]:
            raise IntegrityError(f"{tag}: published SHA-256 metadata changed: {name}")
        if asset.get("size") != pin["size"]:
            raise IntegrityError(f"{tag}: published asset size changed: {name}")
    return actual_assets


def verify_download(repo, tag, name, expected):
    """Hash actual downloaded bytes, not only GitHub's asset digest metadata."""
    url = (
        f"https://github.com/{repo}/releases/download/"
        f"{quote(tag, safe='')}/{quote(name, safe='')}"
    )
    req = Request(url, headers={"User-Agent": "AuronQ-release-integrity-watch/1"})
    digest = hashlib.sha256()
    length = 0
    with urlopen(req, timeout=120) as response:
        while True:
            chunk = response.read(1024 * 1024)
            if not chunk:
                break
            length += len(chunk)
            if length > expected["size"]:
                raise IntegrityError(f"{tag}/{name}: download exceeds pinned size")
            digest.update(chunk)
    if length != expected["size"] or digest.hexdigest() != expected["sha256"]:
        raise IntegrityError(f"{tag}/{name}: DOWNLOADED BYTES DO NOT MATCH PINNED SHA-256")
    print(f"  verified downloaded SHA-256: {name} ({length} bytes)", flush=True)


def run(baseline, metadata_only=False):
    data = manifest_load(baseline)
    repo = data["repository"]
    for expected in data["releases"]:
        tag = expected["tag"]
        print(f"Checking pinned release {tag} ...", flush=True)
        ref = github_json(repo, "git/ref/tags/" + quote(tag, safe=""))
        release = github_json(repo, "releases/tags/" + quote(tag, safe=""))
        assets = verify_metadata(expected, ref, release)
        print(f"  tag, release ID and {len(assets)} asset metadata records match", flush=True)
        if not metadata_only:
            for name in sorted(expected["assets"]):
                verify_download(repo, tag, name, expected["assets"][name])
    print("PASS: all published release identities and pins match.", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--baseline", default=".github/release-integrity-baseline.json"
    )
    parser.add_argument(
        "--metadata-only", action="store_true",
        help="Check GitHub API metadata without downloading asset bytes",
    )
    args = parser.parse_args()
    try:
        run(args.baseline, metadata_only=args.metadata_only)
    except (IntegrityError, OSError, ValueError, KeyError, TypeError) as exc:
        print("RELEASE INTEGRITY CHECK FAILED: " + str(exc), file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
