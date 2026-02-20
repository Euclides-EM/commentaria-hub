import {
  PageLayout,
  SpecialZoomLevel,
  Viewer,
  Worker,
} from "@react-pdf-viewer/core";
import { defaultLayoutPlugin } from "@react-pdf-viewer/default-layout";

import "@react-pdf-viewer/core/lib/styles/index.css";
import "@react-pdf-viewer/default-layout/lib/styles/index.css";

const pageLayout: PageLayout = {
  transformSize: ({ size }) => ({
    height: size.height + 10,
    width: size.width + 10,
  }),
};

export const Presentation = () => {
  const defaultLayoutPluginInstance = defaultLayoutPlugin();
  return (
    <Worker workerUrl="https://unpkg.com/pdfjs-dist@3.4.120/build/pdf.worker.min.js">
      <Viewer
        theme="dark"
        pageLayout={pageLayout}
        defaultScale={SpecialZoomLevel.PageFit}
        fileUrl="/presentation/presentation-jun-25.pdf"
        plugins={[defaultLayoutPluginInstance]}
      />
    </Worker>
  );
};
