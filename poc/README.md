* Downloaded the pdf
* Created jpgs Rescribe GUI (I should just use a script, this is their script https://github.com/rescribe/utils/blob/10eb175a5f651748a57297b35d86a3c9c0987e80/cmd/dlgbook/main.go, it uses getbook, which can be downloaded from https://rescribe.xyz/rescribe/embeds/getgbook-darwin-b14f62f.zip (see ref https://github.com/rescribe/bookpipeline/blob/master/cmd/rescribe/getembeds.go))
* I downloaded the model from https://github.com/PonteIneptique/YALTAi/blob/main/tests/nano-yolo-ladas.pt, according to their article the nano version is quite good (they have a big model, they say the difference is minor).
* I downloaded the ocr model from
* I used only 0013.jpg, and run:
```
yaltai kraken --device cpu -I 0013.jpg --alto  --suffix ".xml" segment --yolo nano-yolo-ladas.pt ocr -m Gallicorpor.mlmodel
```
To run it on a complete folder:
```
yaltai kraken --device cpu -I "/Users/mia/Downloads/ocr_test_8oct/1598_ERRARD_LessixpremiersliuresdeselemensdEuclidetraduictsc_bW9q_TRw0mcC/*.jpg" --alto --suffix ".xml" segment --yolo ../tmp/nano-yolo-ladas.pt ocr -m ../tmp/Gallicorpor.mlmodel
```
* I (AKA ChatGPT) wrote a script to convert the alto to (a) image (alto_render.py) and (b) html (alto_render_html.py).
* I ran the script to create the HTML and image:
```
python alto_render.py 0013.xml 0013.jpg 0013-seg.jpg
python alto_render_html.py 0013.xml 0013.jpg 0013.html
```
* I opened the HTML in a browser, and it looks good.
* I downloaded escriptorium https://gitlab.com/scripta/escriptorium/-/wikis/docker-install
* To run it locally, I had to change the docker-compose.yml. Instead of `image: registry.gitlab.com/scripta/escriptorium/nginx` use `image: escriptorium-nginx:local`
* I followed the instructions here: https://gitlab.com/scripta/escriptorium/-/wikis/docker-install - for running it after the 1st time, just:
```
docker-compose up -d
```
The default user is admin/admin, as specified in variables.env
* I used it with the default segmentation and with the yaltai segmentation and wow the yaltai is so much better.

Convert PDF to JPGs:
```
python pdf_to_jpg.py /Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1.pdf /Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1_jpg
```
Run the yaltai kraken command on a folder:
```
yaltai kraken --device cpu -I "/Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1_jpg/*.jpg" --alto --suffix ".xml" segment --yolo ../tmp/nano-yolo-ladas.pt ocr -m ../tmp/Gallicorpor.mlmodel
```
Fix the XMLs
```
python /Users/mia/dev/personal/elements-dh/tmp/fix_filename_alto.py '/Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1_jpg' -o '/Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1_jpg/fixed'
```
Add METS.xml file to the folder
```
python /Users/mia/dev/personal/elements-dh/tmp/construct_mets_file.py /Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1_jpg/fixed
```
ZIP the folder
```
cd /Users/mia/dev/personal/elements-dh/poc/Paris_1516_book_1_jpg/fixed && zip -r ../Paris_1516_book_1_jpg_fixed.zip *
```
Then upload the zip file in eScriptorium (import -> import from zip file).

Convert to TEI:
clone this repo: https://github.com/TEI4HTR/alto2tei
Then run:
```
xsltproc -o out_tei.xml alto_to_tei.xsl input_alto.xml
```