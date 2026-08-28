import { useContext, useEffect, useMemo, useState } from "react";
import { MAIN_CONTENT_ID } from "../components/layout/routes.ts";
import { useSearchParams } from "react-router-dom";
import styled from "@emotion/styled";
import {
  Container,
  LazyImage,
  Row,
  ScrollToTopButton,
} from "../components/common";
import { Item } from "../types";
import { ItemInfo } from "../components/tps/modal/ItemInfo";
import { NO_CITY, NO_EDITOR, NO_YEAR } from "../constants";
import { joinArr } from "../utils/util.ts";
import { fetchDiagrams, VolumeData } from "../api/diagramsApi.ts";
import { LAND_COLOR, SEA_COLOR } from "../utils/colors.ts";
import { useQuery } from "@tanstack/react-query";
import { isNil } from "lodash";
import {
  FacsimilesService,
  type model_Edition,
  type model_Facsimile,
} from "@hub-api";
import { AuthContext } from "../contexts/Auth.ts";
import { openAuthenticatedFacsimilePDF } from "../utils/facsimilePdf.ts";
import { getEdition } from "../api/editionApi.ts";
import { mapEditionsToItems } from "../utils/dataUtils.ts";

const DiagramsContainer = styled.div`
  max-width: 80vw;
  margin: 0 auto;
  color: white;
`;

const DocumentTitle = styled.h2`
  font-size: 1.5rem;
  font-weight: 600;
`;

const DocumentDescription = styled.div`
  color: #6b7280;
  padding-bottom: 0.5rem;
`;

const VolumeHeader = styled.h3`
  font-size: 1.25rem;
  font-weight: 600;
  color: white;
  margin: 2rem 0 1rem 0;
  border-bottom: 2px solid ${SEA_COLOR};
  padding-bottom: 0.5rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: opacity 0.2s;

  &:hover {
    opacity: 0.8;
  }
`;

const CollapseIcon = styled.span<{ isCollapsed: boolean }>`
  font-size: 0.9rem;
  color: ${LAND_COLOR};
  transition: transform 0.2s;
  transform: ${({ isCollapsed }) =>
    isCollapsed ? "rotate(-90deg)" : "rotate(0deg)"};
  user-select: none;
`;

const DiagramsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
`;

const DiagramCard = styled.div`
  background-color: #eaeaea;
  border-radius: 0.5rem;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.3s;
  display: flex;
  flex-direction: column;
  align-items: center;

  img {
    background-color: #eaeaea;
  }
`;

const DiagramTitle = styled.div`
  font-size: 1.125rem;
  font-weight: 500;
  color: black;
  margin-bottom: 0.25rem;
  padding: 0.5rem;
  span {
    font-size: 0.8rem;
    color: ${SEA_COLOR};
  }
`;

const Modal = styled.div<{ isOpen: boolean }>`
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.75);
  z-index: 50;
  display: ${({ isOpen }) => (isOpen ? "flex" : "none")};
  align-items: center;
  justify-content: center;
  padding: 1rem;
`;

const ModalContent = styled.div`
  background-color: white;
  border-radius: 0.5rem;
  max-width: 64rem;
  max-height: 100%;
  overflow: auto;
`;

const ModalHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid #e5e7eb;
`;

const ModalTitleRow = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
`;

const ModalTitle = styled.h3`
  font-size: 1.125rem;
  font-weight: 500;
  color: #111827;
`;

const ScanPageButton = styled.button`
  padding: 0.4rem 0.75rem;
  background-color: ${SEA_COLOR};
  color: white;
  border: none;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  cursor: pointer;
  white-space: nowrap;

  &:hover {
    opacity: 0.9;
  }
`;

const CloseButton = styled.button`
  color: #9ca3af;
  font-size: 1.5rem;
  font-weight: bold;
  cursor: pointer;
  background: none;
  border: none;

  &:hover {
    color: #6b7280;
  }
`;

const ModalImageContainer = styled.div`
  padding: 1rem;
`;

const ModalImage = styled.img`
  max-width: 100%;
  height: auto;
`;

const NoResults = styled.div`
  grid-column: 1 / -1;
  text-align: center;
  color: #6b7280;
  padding: 2rem;
`;

const FilterContainer = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
  margin: 1rem 0;
  padding: 1rem;
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 0.5rem;
`;

const FilterLabel = styled.label`
  color: white;
  font-size: 0.875rem;
  font-weight: 500;
`;

const FilterInput = styled.input`
  padding: 0.25rem 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  width: 4rem;
  background-color: white;
  color: black;

  &:focus {
    outline: none;
    border-color: ${SEA_COLOR};
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
  }
`;

const FilterButton = styled.button`
  padding: 0.25rem 0.75rem;
  background-color: ${SEA_COLOR};
  color: white;
  border: none;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  cursor: pointer;
  transition: background-color 0.2s;

  &:hover {
    background-color: #2563eb;
  }
`;

interface ImageInfo {
  pageNumber: string;
  index: string;
}

interface ModalState {
  isOpen: boolean;
  imagePath: string;
  title: string;
  scanKey: string;
  pageNumber: number | null;
}

const parseImageName = (imagePath: string): ImageInfo => {
  const filename =
    imagePath.split("/").pop()?.split("?")[0]?.split("#")[0] || imagePath;
  const nameWithoutExt = filename.replace(".jpg", "");
  const parts = nameWithoutExt.split("_");

  if (parts.length >= 3) {
    const pageNumber = parts[0];
    const index = parts[parts.length - 1];

    return {
      pageNumber,
      index,
    };
  }

  return {
    pageNumber: "",
    index: "",
  };
};

function getImagePath(imageName: string) {
  return `${new URL(import.meta.env.VITE_BACKEND_URL).origin}${imageName}`;
}

export const Diagrams = () => {
  const { token } = useContext(AuthContext);
  const [searchParams] = useSearchParams();
  const editionKey = searchParams.get("key");
  const facsimileId = searchParams.get("facsimileId");
  const [collapsedVolumes, setCollapsedVolumes] = useState<Set<string>>(
    new Set(),
  );
  const [showScrollTop, setShowScrollTop] = useState(false);
  const [modal, setModal] = useState<ModalState>({
    isOpen: false,
    imagePath: "",
    title: "",
    scanKey: "",
    pageNumber: null,
  });
  const [pageRangeFrom, setPageRangeFrom] = useState<string>("");
  const [pageRangeTo, setPageRangeTo] = useState<string>("");

  const editionQuery = useQuery({
    queryKey: ["edition", editionKey],
    queryFn: () => getEdition(editionKey!),
    enabled: Boolean(editionKey),
    retry: false,
  });

  const item = useMemo<Item | null>(
    () =>
      editionQuery.data
        ? (mapEditionsToItems([editionQuery.data as model_Edition])[0] ?? null)
        : null,
    [editionQuery.data],
  );

  const diagramsQuery = useQuery({
    queryKey: ["diagrams", editionKey, facsimileId],
    queryFn: () => fetchDiagrams(editionKey || "", facsimileId),
    enabled: Boolean(editionKey || facsimileId),
  });

  const diagramsData = diagramsQuery.data;
  const volumes = useMemo<VolumeData[]>(
    () => diagramsData?.volumes || [],
    [diagramsData?.volumes],
  );
  const images = useMemo<string[]>(
    () => (diagramsData?.volumes ? [] : diagramsData?.images || []),
    [diagramsData?.images, diagramsData?.volumes],
  );
  const scanKeys = useMemo(
    () =>
      Array.from(
        new Set(
          volumes.length > 0
            ? volumes.map((volume) => volume.key).filter(Boolean)
            : editionKey
              ? [editionKey]
              : [],
        ),
      ),
    [editionKey, volumes],
  );
  const facsimilesQuery = useQuery({
    queryKey: ["facsimiles", "download-available", scanKeys],
    queryFn: () => FacsimilesService.getFacsimilies({ editionId: scanKeys }),
    enabled: scanKeys.length > 0,
  });
  const downloadableFacsimilesByScanKey = useMemo(() => {
    const byScanKey = new Map<string, model_Facsimile[]>();
    for (const facsimile of facsimilesQuery.data || []) {
      if (
        !facsimile.download_available ||
        !facsimile.edition_id ||
        !facsimile.id
      ) {
        continue;
      }
      byScanKey.set(facsimile.edition_id, [
        ...(byScanKey.get(facsimile.edition_id) ?? []),
        facsimile,
      ]);
    }
    return byScanKey;
  }, [facsimilesQuery.data]);
  const loading = diagramsQuery.isLoading || editionQuery.isLoading;
  const error =
    (editionQuery.isError ? "Edition not found" : null) ||
    diagramsData?.error ||
    null;

  const openImageModal = (
    imagePath: string,
    image: ImageInfo,
    volumeNumber?: number,
    scanKey = editionKey || "",
  ) => {
    const pageNumber = Number.parseInt(image.pageNumber, 10);
    setModal({
      isOpen: true,
      imagePath,
      title: volumeNumber
        ? `Volume ${volumeNumber} - Page ${image.pageNumber} - Diagram #${image.index}`
        : `Page ${image.pageNumber} - Diagram #${image.index}`,
      scanKey,
      pageNumber: Number.isFinite(pageNumber) ? pageNumber : null,
    });
  };

  const closeImageModal = () => {
    setModal({
      isOpen: false,
      imagePath: "",
      title: "",
      scanKey: "",
      pageNumber: null,
    });
  };

  const openScanPage = (facsimile: model_Facsimile) => {
    if (!token || modal.pageNumber === null) {
      return;
    }
    void openAuthenticatedFacsimilePDF(
      facsimile.id!,
      token,
      modal.pageNumber,
      facsimile.name ? `${facsimile.name}.pdf` : undefined,
    ).catch((error) => {
      console.error("Failed to open scan page:", error);
    });
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        closeImageModal();
      }
    };

    if (modal.isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }
  }, [modal.isOpen]);

  useEffect(() => {
    const el = document.getElementById(MAIN_CONTENT_ID);
    el?.addEventListener("scroll", handleScroll);
    return () => el?.removeEventListener("scroll", handleScroll);
  }, []);

  const filterImagesByPageRange = (images: string[]): string[] => {
    if (!pageRangeFrom && !pageRangeTo) {
      return images;
    }

    const fromPage = pageRangeFrom ? parseInt(pageRangeFrom, 10) : null;
    const toPage = pageRangeTo ? parseInt(pageRangeTo, 10) : null;

    return images.filter((imageName) => {
      const imageInfo = parseImageName(imageName);
      const pageNum = parseInt(imageInfo.pageNumber, 10);

      if (isNaN(pageNum)) return false;

      if (fromPage && pageNum < fromPage) return false;
      if (toPage && pageNum > toPage) return false;

      return true;
    });
  };

  const getAllImages = () => {
    if (volumes.length > 0) {
      return volumes.flatMap((volume) => volume.images);
    }
    return images;
  };

  const getAllFilteredImages = () => {
    return sortImagesByPageNumber(filterImagesByPageRange(getAllImages()));
  };

  const clearPageFilter = () => {
    setPageRangeFrom("");
    setPageRangeTo("");
  };

  const toggleVolumeCollapse = (volumeKey: string) => {
    setCollapsedVolumes((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(volumeKey)) {
        newSet.delete(volumeKey);
      } else {
        newSet.add(volumeKey);
      }
      return newSet;
    });
  };

  const handleScroll = () => {
    const el = document.getElementById(MAIN_CONTENT_ID);
    const scrollTop = el ? el.scrollTop : 0;
    setShowScrollTop(scrollTop > 200);
  };

  const scrollToTop = () => {
    document.getElementById(MAIN_CONTENT_ID)?.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  const sortImagesByPageNumber = (images: string[]): string[] => {
    return images.sort((a, b) => {
      const imageInfoA = parseImageName(a);
      const imageInfoB = parseImageName(b);

      const pageA = parseInt(imageInfoA.pageNumber, 10);
      const pageB = parseInt(imageInfoB.pageNumber, 10);

      if (pageA !== pageB) {
        return pageA - pageB;
      }

      const indexA = parseInt(imageInfoA.index, 10);
      const indexB = parseInt(imageInfoB.index, 10);

      return indexA - indexB;
    });
  };

  const filteredImages = getAllFilteredImages();

  const getPageRange = () => {
    const allImages = getAllImages();
    const pageNumbers = allImages
      .map((imageName) => parseImageName(imageName).pageNumber)
      .filter((pageNumber) => pageNumber !== "")
      .map((pageNumber) => parseInt(pageNumber, 10))
      .filter((pageNum) => !isNaN(pageNum));

    if (pageNumbers.length === 0) {
      return { min: 1, max: 999 };
    }

    return {
      min: Math.min(...pageNumbers),
      max: Math.max(...pageNumbers),
    };
  };

  const pageRange = getPageRange();

  if (!editionKey) {
    return (
      <Container>
        <DiagramsContainer>
          <DocumentTitle>Document Not Found</DocumentTitle>
          <DocumentDescription>
            The requested document could not be found.
          </DocumentDescription>
        </DiagramsContainer>
      </Container>
    );
  }

  return (
    <>
      {showScrollTop && (
        <ScrollToTopButton onClick={scrollToTop} title="Scroll to top">
          ↑
        </ScrollToTopButton>
      )}
      <DiagramsContainer>
        {item ? (
          <DocumentTitle>
            {item.year || NO_YEAR} {joinArr(item.editors) || NO_EDITOR},{" "}
            {joinArr(item.cities) || NO_CITY}
          </DocumentTitle>
        ) : (
          <DocumentTitle>Loading...</DocumentTitle>
        )}

        {item && <ItemInfo item={item} showDiagramsLink={false} isRow />}

        <DocumentDescription>
          {loading
            ? "Loading diagrams..."
            : error
              ? error
              : getAllImages().length === 0
                ? "No diagrams detected."
                : (() => {
                    const allImages = getAllImages();
                    const distinctPages = new Set(
                      allImages
                        .map(
                          (imageName) => parseImageName(imageName).pageNumber,
                        )
                        .filter((pageNumber) => pageNumber !== ""),
                    ).size;
                    const filteredCount = filteredImages.length;
                    const filteredDistinctPages = new Set(
                      filteredImages
                        .map(
                          (imageName) => parseImageName(imageName).pageNumber,
                        )
                        .filter((pageNumber) => pageNumber !== ""),
                    ).size;

                    const volumeText =
                      volumes.length > 1
                        ? `, across ${volumes.length} volumes`
                        : "";

                    if (pageRangeFrom || pageRangeTo) {
                      return `${filteredCount.toLocaleString()} diagram${filteredCount === 1 ? "" : "s"} shown across ${filteredDistinctPages.toLocaleString()} distinct page${filteredDistinctPages === 1 ? "" : "s"} (${allImages.length} total${volumeText})`;
                    }

                    return `${allImages.length.toLocaleString()} diagram${allImages.length === 1 ? "" : "s"} detected across ${distinctPages.toLocaleString()} distinct page${distinctPages === 1 ? "" : "s"}${volumeText}`;
                  })()}
        </DocumentDescription>

        {item && (
          <Row justifyStart>
            {item.hasDiagrams && !isNil(item.hasDiagrams) && (
              <DocumentDescription>
                <strong>Has Diagrams:</strong> {item.hasDiagrams}
              </DocumentDescription>
            )}
          </Row>
        )}

        {!loading && !error && getAllImages().length > 0 && (
          <FilterContainer>
            <FilterLabel>Filter by page range:</FilterLabel>
            <FilterLabel>From:</FilterLabel>
            <FilterInput
              type="number"
              value={pageRangeFrom}
              onChange={(e) => setPageRangeFrom(e.target.value)}
              placeholder={pageRange.min.toString()}
              min={pageRange.min}
              max={pageRange.max}
            />
            <FilterLabel>To:</FilterLabel>
            <FilterInput
              type="number"
              value={pageRangeTo}
              onChange={(e) => setPageRangeTo(e.target.value)}
              placeholder={pageRange.max.toString()}
              min={pageRange.min}
              max={pageRange.max}
            />
            {(pageRangeFrom || pageRangeTo) && (
              <FilterButton onClick={clearPageFilter}>Clear</FilterButton>
            )}
          </FilterContainer>
        )}

        {volumes.length > 0 ? (
          volumes.map((volume) => {
            const volumeImages = sortImagesByPageNumber(
              filterImagesByPageRange(volume.images),
            );
            const isCollapsed = collapsedVolumes.has(volume.key);

            return (
              <div key={volume.key}>
                {volume.volume && (
                  <VolumeHeader
                    onClick={() => toggleVolumeCollapse(volume.key)}
                  >
                    <span>Volume {volume.volume}</span>
                    <CollapseIcon isCollapsed={isCollapsed}>▼</CollapseIcon>
                  </VolumeHeader>
                )}
                {!isCollapsed && (
                  <DiagramsGrid>
                    {volumeImages.length === 0 && volume.images.length > 0 && (
                      <NoResults>
                        No diagrams found in the specified page range for this
                        volume.
                      </NoResults>
                    )}
                    {volumeImages.map((imageName) => {
                      const imageInfo = parseImageName(imageName);
                      const imagePath = getImagePath(imageName);

                      return (
                        <DiagramCard
                          key={`${volume.key}-${imageName}`}
                          onClick={() =>
                            openImageModal(
                              imagePath,
                              imageInfo,
                              volume.volume,
                              volume.key,
                            )
                          }
                        >
                          <LazyImage
                            src={imagePath}
                            alt={`${imageInfo.pageNumber} - Diagram ${imageInfo.index}`}
                            height="12rem"
                            placeholder="Loading diagram..."
                          />
                          <DiagramTitle>
                            Page {imageInfo.pageNumber}{" "}
                            <span>#{imageInfo.index}</span>
                          </DiagramTitle>
                        </DiagramCard>
                      );
                    })}
                  </DiagramsGrid>
                )}
              </div>
            );
          })
        ) : (
          <DiagramsGrid>
            {!loading &&
              !error &&
              filteredImages.length === 0 &&
              images.length > 0 && (
                <NoResults>
                  No diagrams found in the specified page range.
                </NoResults>
              )}

            {filteredImages.map((imageName) => {
              const imageInfo = parseImageName(imageName);
              const imagePath = getImagePath(imageName);

              return (
                <DiagramCard
                  key={imageName}
                  onClick={() => openImageModal(imagePath, imageInfo)}
                >
                  <LazyImage
                    src={imagePath}
                    alt={`${imageInfo.pageNumber} - Diagram ${imageInfo.index}`}
                    height="12rem"
                    placeholder="Loading diagram..."
                  />
                  <DiagramTitle>
                    Page {imageInfo.pageNumber} <span>#{imageInfo.index}</span>
                  </DiagramTitle>
                </DiagramCard>
              );
            })}
          </DiagramsGrid>
        )}

        <Modal isOpen={modal.isOpen} onClick={closeImageModal}>
          <ModalContent onClick={(e) => e.stopPropagation()}>
            <ModalHeader>
              <ModalTitleRow>
                <ModalTitle>{modal.title}</ModalTitle>
                {token &&
                  modal.pageNumber !== null &&
                  (
                    downloadableFacsimilesByScanKey.get(modal.scanKey) ?? []
                  ).map((facsimile) => (
                    <ScanPageButton
                      key={facsimile.id}
                      type="button"
                      onClick={() => openScanPage(facsimile)}
                    >
                      {facsimile.name
                        ? `View page in ${facsimile.name}`
                        : "View page in scan"}
                    </ScanPageButton>
                  ))}
              </ModalTitleRow>
              <CloseButton onClick={closeImageModal}>×</CloseButton>
            </ModalHeader>
            <ModalImageContainer>
              <ModalImage src={modal.imagePath} alt={modal.title} />
            </ModalImageContainer>
          </ModalContent>
        </Modal>
      </DiagramsContainer>
    </>
  );
};
