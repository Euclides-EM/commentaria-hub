import { useState, memo, lazy, Suspense, useRef, useEffect } from "react";
import { ItemProps } from "../../../types";
import {
  Column,
  ExpandIcon,
  ImageExpandIcon,
  ImageTile,
  LanguagesInfo,
  NoImageTile,
  StyledImage,
  TextTile,
} from "../../common.ts";
import { openScan } from "../../../utils/dataUtils.ts";
import { ItemModal } from "../modal/ItemModal.tsx";
import { NO_AUTHOR, NO_CITY, NO_YEAR } from "../../../constants";
import { joinArr, toItemImageUrl } from "../../../utils/util.ts";
import styled from "@emotion/styled";

const HighlightedText = lazy(() =>
  import("./HighlightedText.tsx").then((module) => ({
    default: module.HighlightedText,
  })),
);

const NoTitlePage = styled.div`
  flex: 1;
  text-align: center;
  color: darkgray;
  margin-top: calc(50% - 2.5rem);
`;

export const ItemView = memo(
  ({ item, height, width, mode, featuresById }: ItemProps) => {
    const [modalOpen, setModalOpen] = useState(false);
    const [isVisible, setIsVisible] = useState(false);
    const itemRef = useRef<HTMLDivElement>(null);
    const imageUrl = toItemImageUrl(item.tpImageName);

    useEffect(() => {
      const observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting) {
            setIsVisible(true);
            observer.disconnect();
          }
        },
        { threshold: 0.1 },
      );

      if (itemRef.current) {
        observer.observe(itemRef.current);
      }

      return () => {
        observer.disconnect();
      };
    }, []);

    return (
      <Column id={item.key} style={{ height, width }} ref={itemRef}>
        <div style={{ zIndex: 11 }}>
          {item.year || NO_YEAR} {joinArr(item.authors) || NO_AUTHOR},{" "}
          {joinArr(item.cities) || NO_CITY}
          <LanguagesInfo>{joinArr(item.languages)}</LanguagesInfo>
        </div>
        {mode === "texts" && (
          <>
            {item.title === "?" ? (
              <>
                <NoImageTile>
                  No title page available
                  <ExpandIcon title="Expand" onClick={() => setModalOpen(true)}>
                    ⤢
                  </ExpandIcon>
                </NoImageTile>
              </>
            ) : (
              <>
                <TextTile alignCenter={!!imageUrl}>
                  {!isVisible ? (
                    <div>
                      {item.title}
                      {item.imprint && (
                        <>
                          <hr style={{ opacity: 0.3 }} />
                          {item.imprint}
                        </>
                      )}
                    </div>
                  ) : (
                    <Suspense
                      fallback={
                        <div>
                          {item.title}
                          {item.imprint && (
                            <>
                              <hr style={{ opacity: 0.3 }} />
                              {item.imprint}
                            </>
                          )}
                        </div>
                      }
                    >
                      {item.title && (
                        <HighlightedText
                          text={item.title}
                          featuresById={featuresById || {}}
                          itemKey={item.key}
                        />
                      )}
                      {item.imprint && (
                        <>
                          <hr style={{ opacity: 0.3 }} />
                          {item.imprint}
                        </>
                      )}
                      {!item.title && (
                        <NoTitlePage>
                          This edition has no title page or it is not available.
                        </NoTitlePage>
                      )}
                    </Suspense>
                  )}
                  <ExpandIcon title="Expand" onClick={() => setModalOpen(true)}>
                    ⤢
                  </ExpandIcon>
                </TextTile>
              </>
            )}
          </>
        )}
        {mode === "images" &&
          (item.tpImageName ? (
            <ImageTile>
              <StyledImage
                src={toItemImageUrl(item.tpImageName)!}
                onClick={() => openScan(item)}
                clickable={!!item.scanUrl && item.scanUrl.length > 0}
              />
              <ImageExpandIcon
                title="Expand"
                onClick={(e) => {
                  e.stopPropagation();
                  setModalOpen(true);
                }}
              >
                ⤢
              </ImageExpandIcon>
            </ImageTile>
          ) : (
            <NoImageTile>
              Not Available
              <ExpandIcon title="Expand" onClick={() => setModalOpen(true)}>
                ⤢
              </ExpandIcon>
            </NoImageTile>
          ))}

        {modalOpen && (
          <ItemModal
            item={item}
            featuresById={featuresById}
            onClose={() => setModalOpen(false)}
          />
        )}
      </Column>
    );
  },
);
