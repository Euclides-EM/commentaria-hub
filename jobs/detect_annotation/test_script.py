import os
import unittest
from pathlib import Path
from unittest.mock import patch

import script


class XMLAndMaskSafetyTest(unittest.TestCase):
    def test_atomic_xml_write_preserves_destination_when_serialization_fails(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            path = Path(directory) / "page.xml"
            path.write_text("original", encoding="utf-8")

            class FailingTree:
                def write(self, destination, **_kwargs):
                    Path(destination).write_bytes(b"")
                    raise OSError("simulated write failure")

            with self.assertRaisesRegex(OSError, "simulated write failure"):
                script.write_xml_atomic(FailingTree(), path)

            self.assertEqual(path.read_text(encoding="utf-8"), "original")
            self.assertEqual(list(path.parent.glob(".page.xml.*.tmp")), [])

    def test_create_mask_rejects_region_erased_by_ignore_category(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            root = Path(directory)
            alto_path = root / "page.xml"
            mask_path = root / "mask.png"
            alto_path.write_text(
                '<alto><Tags><OtherTag ID="main" LABEL="MainZone"/>'
                '<OtherTag ID="ignore" LABEL="IgnoreZone"/></Tags><Layout>'
                '<Page WIDTH="100" HEIGHT="100"><PrintSpace>'
                '<TextBlock TAGREFS="main" HPOS="10" VPOS="10" WIDTH="20" HEIGHT="20"/>'
                '<TextBlock TAGREFS="ignore" HPOS="0" VPOS="0" WIDTH="100" HEIGHT="100"/>'
                '</PrintSpace></Page></Layout></alto>',
                encoding="utf-8",
            )

            self.assertFalse(
                script.create_mask(alto_path, mask_path, ["MainZone"], ["IgnoreZone"])
            )
            self.assertEqual(script.Image.open(mask_path).getextrema(), (0, 0))

    def test_create_mask_accepts_finished_bitonal_mask(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            root = Path(directory)
            alto_path = root / "page.xml"
            mask_path = root / "mask.png"
            alto_path.write_text(
                '<alto><Tags><OtherTag ID="main" LABEL="MainZone"/></Tags>'
                '<Layout><Page WIDTH="100" HEIGHT="100"><PrintSpace>'
                '<TextBlock TAGREFS="main" HPOS="10" VPOS="10" WIDTH="20" HEIGHT="20"/>'
                '</PrintSpace></Page></Layout></alto>',
                encoding="utf-8",
            )

            self.assertTrue(script.create_mask(alto_path, mask_path, ["MainZone"], []))
            self.assertEqual(script.Image.open(mask_path).getextrema(), (0, 1))


class FailureCallbackTest(unittest.TestCase):
    def test_upload_failure_posts_mode_error_and_bearer_token(self):
        result = type("Result", (), {"returncode": 0, "stdout": "", "stderr": ""})()
        with patch.object(script.shutil, "which", return_value="/usr/bin/curl"):
            with patch.object(script.subprocess, "run", return_value=result) as run:
                script.upload_failure(
                    "https://example.test/detection_failure",
                    "secret-token",
                    "lines",
                    "Document is empty",
                )

        command = run.call_args.args[0]
        self.assertIn("Authorization: Bearer secret-token", command)
        self.assertIn("mode=lines", command)
        self.assertIn("error=Document is empty", command)
        self.assertEqual(command[-1], "https://example.test/detection_failure")

    def test_main_reports_processing_failure(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            root = Path(directory)
            environment = {
                "MODE": "lines",
                "IMAGE_DIR": str(root / "images"),
                "ALTO_DIR": str(root / "alto"),
                "OUTPUT_DIR": str(root / "output"),
                "ARTIFACTS_DIR": str(root / "artifacts"),
                "RESULT_FAILURE_URL": "https://example.test/detection_failure",
                "RESULT_UPLOAD_TOKEN": "secret-token",
            }
            with patch.dict(os.environ, environment, clear=True):
                with patch.object(script, "detect_lines", side_effect=RuntimeError("boom")):
                    with patch.object(script, "upload_failure") as upload_failure:
                        self.assertEqual(script.main(), 1)

            upload_failure.assert_called_once_with(
                "https://example.test/detection_failure",
                "secret-token",
                "lines",
                "boom",
            )


class WorkerHelpersTest(unittest.TestCase):
    def test_worker_count_uses_slurm_allocation_and_caps_at_jobs(self):
        with patch.dict(os.environ, {"SLURM_CPUS_PER_TASK": "4"}, clear=True):
            self.assertEqual(script.worker_count(10), 4)
            self.assertEqual(script.worker_count(2), 2)
            self.assertEqual(script.worker_count(0), 0)

    def test_explicit_worker_count_wins(self):
        with patch.dict(
            os.environ,
            {"SLURM_CPUS_PER_TASK": "4", "DETECTION_WORKERS": "2"},
            clear=True,
        ):
            self.assertEqual(script.worker_count(10), 2)

    def test_shard_is_round_robin_and_complete(self):
        pages = [Path(f"page-{index}.png") for index in range(7)]
        self.assertEqual(
            script.shard(pages, 3),
            [[pages[0], pages[3], pages[6]], [pages[1], pages[4]], [pages[2], pages[5]]],
        )


class KrakenOCRTest(unittest.TestCase):
    def test_command_maps_default_line_type_to_model(self):
        command = script.kraken_ocr_command(
            ["-i", "/tmp/page.xml", "/tmp/page.ocr.xml"],
            Path("/tmp/model.mlmodel"),
        )
        self.assertEqual(command[-3:], ["ocr", "-m", "default:/tmp/model.mlmodel"])

    def test_prepare_alto_for_ocr_repairs_geometry_and_image_path(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            alto_path = Path(directory) / "page.xml"
            alto_path.write_text(
                '<alto xmlns="http://www.loc.gov/standards/alto/ns-v4#"><Description>'
                '<sourceImageInformation><fileName>old.png</fileName>'
                '</sourceImageInformation></Description><Layout>'
                '<Page WIDTH="100" HEIGHT="100"><PrintSpace HPOS="0" VPOS="0" '
                'WIDTH="100" HEIGHT="100">'
                '<TextBlock ID="empty"/><TextBlock HPOS="0" VPOS="0" WIDTH="100" '
                'HEIGHT="100" TAGREFS="region-main"><TextLine ID="line-1" '
                'TAGREFS="line-default" HPOS="10" VPOS="20" '
                'WIDTH="30" HEIGHT="40" BASELINE="10 50 40 50"/></TextBlock>'
                '</PrintSpace></Page></Layout></alto>',
                encoding="utf-8",
            )

            self.assertTrue(
                script.prepare_alto_for_ocr(alto_path, Path("/images/page.png"))
            )

            tree = script.etree.parse(str(alto_path))
            self.assertEqual(
                script.xpath(tree, "//*[local-name()='fileName']")[0].text,
                "/images/page.png",
            )
            line = script.xpath(tree, "//*[local-name()='TextLine']")[0]
            polygon = script.xpath(
                line, "./*[local-name()='Shape']/*[local-name()='Polygon']"
            )[0]
            self.assertEqual(polygon.get("POINTS"), "10 20 40 20 40 60 10 60 10 20")
            self.assertIsNone(line.get("TAGREFS"))
            block = script.xpath(tree, "//*[local-name()='TextBlock' and @TAGREFS]")[0]
            self.assertEqual(block.get("TAGREFS"), "region-main")
            self.assertFalse(script.xpath(tree, "//*[@ID='empty']"))

            from kraken.lib.xml import XMLPage

            segmentation = XMLPage(alto_path, filetype="alto").to_container()
            self.assertTrue(segmentation.script_detection)
            self.assertEqual(segmentation.lines[0].tags, {"type": "default"})

    def test_prepare_alto_for_ocr_skips_blank_page(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            alto_path = Path(directory) / "page.xml"
            alto_path.write_text(
                '<alto><Description><sourceImageInformation><fileName>old.png</fileName>'
                '</sourceImageInformation></Description><Layout><Page><PrintSpace>'
                '<TextBlock HPOS="0" VPOS="0" WIDTH="100" HEIGHT="100"/>'
                '</PrintSpace></Page></Layout></alto>',
                encoding="utf-8",
            )

            self.assertFalse(
                script.prepare_alto_for_ocr(alto_path, Path("/images/page.png"))
            )

    def test_model_ocr_does_not_launch_kraken_for_blank_page(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            root = Path(directory)
            image_dir = root / "images"
            alto_dir = root / "alto"
            output_dir = root / "output"
            image_dir.mkdir()
            alto_dir.mkdir()
            (image_dir / "page-0001.png").write_bytes(b"")
            input_alto = (
                '<alto><Description><sourceImageInformation><fileName>page-0001.png'
                '></fileName></sourceImageInformation></Description><Layout><Page>'
                '<PrintSpace><TextBlock HPOS="0" VPOS="0" WIDTH="100" HEIGHT="100"/>'
                '</PrintSpace></Page></Layout></alto>'
            )
            (alto_dir / "page-0001.xml").write_text(input_alto, encoding="utf-8")

            with patch.object(script, "run") as run:
                script.model_ocr(image_dir, alto_dir, output_dir, Path("/missing/model.mlmodel"))

            run.assert_not_called()
            self.assertEqual(
                (output_dir / "page-0001.xml").read_text(encoding="utf-8"),
                input_alto,
            )

    def test_model_ocr_shards_pages_across_allocated_workers(self):
        from tempfile import TemporaryDirectory

        with TemporaryDirectory() as directory:
            root = Path(directory)
            image_dir = root / "images"
            alto_dir = root / "alto"
            output_dir = root / "output"
            image_dir.mkdir()
            alto_dir.mkdir()
            for page in range(1, 5):
                stem = f"page-{page:04d}"
                (image_dir / f"{stem}.png").write_bytes(b"")
                (alto_dir / f"{stem}.xml").write_text(
                    '<alto><Description><sourceImageInformation><fileName>'
                    f'{stem}.png</fileName></sourceImageInformation></Description>'
                    '<Layout><Page><PrintSpace><TextBlock HPOS="0" VPOS="0" '
                    'WIDTH="100" HEIGHT="100"><TextLine ID="line" BASELINE="0 50 100 50" '
                    'HPOS="0" VPOS="0" WIDTH="100" HEIGHT="100"/></TextBlock>'
                    '</PrintSpace></Page></Layout></alto>',
                    encoding="utf-8",
                )

            def fake_kraken(command: list[str]) -> None:
                for index, argument in enumerate(command):
                    if argument == "-i":
                        script.shutil.copyfile(command[index + 1], command[index + 2])

            with patch.dict(os.environ, {"SLURM_CPUS_PER_TASK": "2"}, clear=True):
                with patch.object(script, "run", side_effect=fake_kraken) as run:
                    script.model_ocr(
                        image_dir,
                        alto_dir,
                        output_dir,
                        Path("/models/model.mlmodel"),
                    )

            self.assertEqual(run.call_count, 2)
            self.assertTrue(all((output_dir / f"page-{page:04d}.xml").exists() for page in range(1, 5)))


if __name__ == "__main__":
    unittest.main()
