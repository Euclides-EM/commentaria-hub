import { Item } from "../../../types";
import { StyledImage } from "../../common.ts";
import {
  Modal,
  ModalClose,
  ModalContent,
  ModalTextColumn,
  ModalTextContainer,
  ModalTitle,
  TextColumnsContainer,
} from "./ModalComponents.tsx";
import { lazy, Suspense, useContext } from "react";
import { openImage } from "../../../utils/dataUtils.ts";
import styled from "@emotion/styled";
import { HelpTip } from "../../map/Filter.tsx";
import {
  TOOLTIP_EN_TRANSLATION,
  TOOLTIP_TRANSCRIPTION,
} from "../../map/MapTooltips.tsx";
import { NotesEditor } from "./NotesEditor.tsx";
import { ItemInfo } from "./ItemInfo.tsx";
import { toItemImageUrl } from "../../../utils/util.ts";
import { feature_Feature } from "@hub-api";
import { AuthContext } from "../../../contexts/Auth.ts";

const HighlightedText = lazy(() =>
  import("../features/HighlightedText.tsx").then((module) => ({
    default: module.HighlightedText,
  })),
);

type ItemModalProps = {
  item: Item;
  featuresById: Record<string, feature_Feature> | null;
  onClose: () => void;
};

const StyledHelpTip = styled(HelpTip)`
  margin: 0 0 0 -0.5rem;
  z-index: 100;
  svg {
    margin-bottom: 4px;
  }
`;

const NoTitlePage = styled.div`
  flex: 1;
  text-align: center;
  color: darkgray;
`;

export const ItemModal = ({ item, featuresById, onClose }: ItemModalProps) => {
  const highlightFeatures = featuresById || {};
  const hasTitleText = !!item.title && item.title !== "?";
  const imageUrl = toItemImageUrl(item.tpImageName);
  const { token } = useContext(AuthContext);

  return (
    <Modal onClick={onClose}>
      <ModalContent
        onClick={(e) => e.stopPropagation()}
        hasImage={!!item.tpImageName}
      >
        <ModalClose title="Close" onClick={onClose}>
          ✕
        </ModalClose>
        <ItemInfo isRow={Boolean(imageUrl || hasTitleText)} item={item} />
        <ModalTextContainer>
          {imageUrl && (
            <ModalTextColumn isImage>
              <StyledImage
                large
                clickable
                src={imageUrl}
                onClick={() => openImage(item)}
              />
            </ModalTextColumn>
          )}
          {hasTitleText ? (
            <TextColumnsContainer>
              <ModalTextColumn isTextContent alignCenter={!!imageUrl}>
                <ModalTitle justifyStart gap={1}>
                  Original Text
                  <StyledHelpTip tooltipId={TOOLTIP_TRANSCRIPTION} />
                </ModalTitle>
                <Suspense fallback={<div>{item.title}</div>}>
                  <HighlightedText
                    text={item.title!}
                    featuresById={highlightFeatures}
                    itemKey={item.key}
                  />
                </Suspense>
                {item.imprint && (
                  <>
                    <hr style={{ opacity: 0.3 }} />
                    {item.imprint}
                  </>
                )}
              </ModalTextColumn>
              {(item.titleEn || item.imprintEn) && (
                <ModalTextColumn isTextContent alignCenter={!!imageUrl}>
                  <ModalTitle justifyStart gap={1}>
                    English Translation{" "}
                    <StyledHelpTip tooltipId={TOOLTIP_EN_TRANSLATION} />
                  </ModalTitle>
                  <div>{item.titleEn}</div>
                  {item.imprintEn && (
                    <>
                      {item.imprint && <hr style={{ opacity: 0.3 }} />}
                      <div>{item.imprintEn}</div>
                    </>
                  )}
                </ModalTextColumn>
              )}
            </TextColumnsContainer>
          ) : (
            <NoTitlePage>
              This edition has no title page or it is not available.
            </NoTitlePage>
          )}
        </ModalTextContainer>
        {token && <NotesEditor item={item} />}
      </ModalContent>
    </Modal>
  );
};
