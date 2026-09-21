import importlib.util
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import numpy as np


SCRIPT_PATH = Path(__file__).with_name("script.py")
SPEC = importlib.util.spec_from_file_location("train_ocr_script", SCRIPT_PATH)
SCRIPT = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(SCRIPT)


class CompileKrakenDatasetTest(unittest.TestCase):
    def test_uses_seeded_random_split(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            manifest = tmp_path / "alto_files.txt"
            manifest.write_text("/tmp/page-0001.xml\n/tmp/page-0002.xml\n", encoding="utf-8")
            draws = []
            calls = []

            def fake_build_binary_dataset(**kwargs):
                calls.append(kwargs)
                draws.append(np.random.random(5))

            with patch("kraken.lib.arrow_dataset.build_binary_dataset", side_effect=fake_build_binary_dataset):
                SCRIPT.compile_kraken_dataset(manifest, tmp_path / "first.arrow", 42)
                SCRIPT.compile_kraken_dataset(manifest, tmp_path / "second.arrow", 42)

            np.testing.assert_array_equal(draws[0], draws[1])
            self.assertEqual(calls[0]["files"], ["/tmp/page-0001.xml", "/tmp/page-0002.xml"])
            self.assertEqual(calls[0]["random_split"], (0.8, 0.1, 0.1))
            self.assertEqual(calls[0]["format_type"], "alto")


if __name__ == "__main__":
    unittest.main()
