"""Negative tests for independent, read-only official release integrity checks."""
import hashlib
import io
import json
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import verify_release_integrity as watch


class PinnedReleaseIntegrityTests(unittest.TestCase):
    def setUp(self):
        self.body = b"known official binary bytes"
        self.meta = {
            "size": len(self.body),
            "sha256": hashlib.sha256(self.body).hexdigest(),
        }
        self.expected = {
            "tag": "v1.7.13",
            "commit": "a" * 40,
            "release_id": 403148314,
            "assets": {"sample.bin": self.meta},
        }
        self.ref = {
            "ref": "refs/tags/v1.7.13",
            "object": {"type": "commit", "sha": "a" * 40},
        }
        self.release = {
            "tag_name": "v1.7.13",
            "id": 403148314,
            "draft": False,
            "published_at": "2026-10-04T18:20:57Z",
            "assets": [
                {"name": "sample.bin", "size": len(self.body),
                 "digest": "sha256:" + self.meta["sha256"]}
            ],
        }

    def test_valid_metadata_matches(self):
        result = watch.verify_metadata(self.expected, self.ref, self.release)
        self.assertEqual(len(result), 1)

    def test_retargeted_tag_rejected(self):
        self.ref["object"]["sha"] = "b" * 40
        with self.assertRaisesRegex(watch.IntegrityError, "tag moved"):
            watch.verify_metadata(self.expected, self.ref, self.release)

    def test_recreated_release_rejected(self):
        self.release["id"] += 1
        with self.assertRaisesRegex(watch.IntegrityError, "recreated"):
            watch.verify_metadata(self.expected, self.ref, self.release)

    def test_replaced_asset_and_extra_asset_rejected(self):
        self.release["assets"][0]["digest"] = "sha256:" + "0" * 64
        with self.assertRaisesRegex(watch.IntegrityError, "SHA-256"):
            watch.verify_metadata(self.expected, self.ref, self.release)
        self.release["assets"][0]["digest"] = "sha256:" + self.meta["sha256"]
        self.release["assets"].append({"name": "unrecognized.exe"})
        with self.assertRaisesRegex(watch.IntegrityError, "inventory"):
            watch.verify_metadata(self.expected, self.ref, self.release)

    def test_downloaded_tampered_bytes_rejected(self):
        with mock.patch.object(watch, "urlopen", return_value=io.BytesIO(self.body)):
            watch.verify_download("promirmir/AuronQ", "v1.7.13", "sample.bin", self.meta)
        altered = b"known official binary byte!"
        self.assertEqual(len(altered), len(self.body))
        with mock.patch.object(watch, "urlopen", return_value=io.BytesIO(altered)):
            with self.assertRaisesRegex(watch.IntegrityError, "DOWNLOADED BYTES"):
                watch.verify_download(
                    "promirmir/AuronQ", "v1.7.13", "sample.bin", self.meta
                )

    def test_baseline_rejects_invalid_digest(self):
        fake = {
            "schema": 1,
            "repository": "promirmir/AuronQ",
            "releases": [dict(self.expected, tag=t) for t in ("a", "b", "c", "d", "e", "f")],
        }
        with tempfile.TemporaryDirectory() as root:
            path = Path(root, "release.json")
            path.write_text(json.dumps(fake), encoding="utf-8")
            self.assertEqual(len(watch.manifest_load(path)["releases"]), 6)
            fake["releases"][0]["assets"]["sample.bin"]["sha256"] = "not-sha256"
            path.write_text(json.dumps(fake), encoding="utf-8")
            with self.assertRaisesRegex(watch.IntegrityError, "invalid pinned SHA"):
                watch.manifest_load(path)


if __name__ == "__main__":
    unittest.main()
