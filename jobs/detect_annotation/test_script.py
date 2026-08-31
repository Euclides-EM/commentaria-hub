import os
import unittest
from pathlib import Path
from unittest.mock import patch

import script


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

            script.prepare_alto_for_ocr(alto_path, Path("/images/page.png"))

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


if __name__ == "__main__":
    unittest.main()
