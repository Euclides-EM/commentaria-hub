import { Tooltip } from "react-tooltip";
import styled from "@emotion/styled";

export const TOOLTIP_FEATURES_HIGHLIGHT = "features-highlight";
export const TOOLTIP_RESET = "reset";
export const TOOLTIP_ZOOMIN = "zoomin";
export const TOOLTIP_ZOOMOUT = "zoomout";
export const TOOLTIP_CLOSE_CITY_DETAILS = "close-city-details";
export const TOOLTIP_CLOSE_RECORD_DETAILS = "close-record-details";
export const TOOLTIP_FILTERS = "filters";
export const TOOLTIP_TIMELINE_BUTTON = "timeline-button";
export const TOOLTIP_HELP = "help";
export const TOOLTIP_MARKER_ARROW = "marker-arrow";
export const TOOLTIP_SCAN = "view-scan";
export const TOOLTIP_BOOK_TYPE = "book-type";
export const TOOLTIP_TRANSCRIPTION = "transcription";
export const TOOLTIP_EN_TRANSLATION = "en-translation";

const LimitedWidthDiv = styled.div`
  max-width: 256px;
  text-align: justify;
`;

export const MapTooltips = () => (
  <>
    <Tooltip id={TOOLTIP_RESET} />
    <Tooltip id={TOOLTIP_ZOOMIN} />
    <Tooltip id={TOOLTIP_ZOOMOUT} />
    <Tooltip id={TOOLTIP_CLOSE_CITY_DETAILS} />
    <Tooltip id={TOOLTIP_CLOSE_RECORD_DETAILS} />
    <Tooltip id={TOOLTIP_FILTERS} />
    <Tooltip id={TOOLTIP_TIMELINE_BUTTON} />
    <Tooltip id={TOOLTIP_HELP} />
    <Tooltip id={TOOLTIP_SCAN} />
    <Tooltip id={TOOLTIP_TRANSCRIPTION}>
      Transcription was constructed in part using an LLM
    </Tooltip>
    <Tooltip id={TOOLTIP_EN_TRANSLATION}>
      Translation was constructed in part using an LLM
    </Tooltip>
    <Tooltip id={TOOLTIP_BOOK_TYPE}>
      The classification of a book as a distinct edition or as a translation of
      Elements is not always clear and can be challenged.
    </Tooltip>
    <Tooltip anchorSelect={`#${TOOLTIP_MARKER_ARROW}`} offset={-14}>
      <LimitedWidthDiv>
        Marker placement is for representation purposes and does not reflect the
        actual northeastern location on the Asian continent.
      </LimitedWidthDiv>
    </Tooltip>
  </>
);
