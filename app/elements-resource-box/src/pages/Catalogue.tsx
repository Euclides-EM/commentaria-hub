import { upperFirst } from "lodash";
import { MAIN_CONTENT_ID } from "../components/layout/routes.ts";
import { useCallback, useContext, useEffect, useMemo, useState } from "react";
import {
  ColumnDef,
  ColumnResizeMode,
  createColumnHelper,
  ExpandedState,
  flexRender,
  getCoreRowModel,
  getExpandedRowModel,
  SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import styled from "@emotion/styled";
import { FacsimilesService, type search_OrderByOption } from "@hub-api";
import { SiMaterialdesign } from "react-icons/si";
import { Item, STUDY_CORPUSES } from "../types";
import { useAppliedFilter } from "../contexts/FilterAppliedContext";
import { useEditionsSearchInfinite } from "../hooks/useEditionsSearch";
import {
  Container,
  Row,
  ScrollbarStyle,
  ScrollToTopButton,
} from "../components/common";
import { ItemModal } from "../components/tps/modal/ItemModal";
import { ItemTypes, NO_CITY } from "../constants";
import { joinArr } from "../utils/util.ts";
import {
  FaCheck,
  FaChevronDown,
  FaChevronRight,
  FaFilePdf,
} from "react-icons/fa";
import { AiFillEdit, AiOutlineCopy } from "react-icons/ai";
import { SEA_COLOR } from "../utils/colors.ts";
import { AuthContext } from "../contexts/Auth.ts";
import { Switch, SwitchOption } from "../components/map/Switch.tsx";
import { Stats } from "../components/Stats.tsx";
import { exportCitationsAsRTF } from "../utils/chicagoCitationExport";
import { useLocalStorage } from "usehooks-ts";
import { withAppBasePath } from "../utils/basePath";
import { PiArrowBendDownRightBold } from "react-icons/pi";
import { inEuclidesMode } from "../utils/mode.ts";
import { useAutoOpenEditionFromQuery } from "../hooks/useAutoOpenEditionFromQuery.ts";
import { FacsimileLinks } from "../components/FacsimileLinks.tsx";
import {
  formatDisplayBooks,
  formatDisplayEditors,
  formatDisplayYear,
} from "../utils/itemDisplay.ts";
import { openAuthenticatedFacsimilePDF } from "../utils/facsimilePdf.ts";

const TableContainer = styled.div`
  ${ScrollbarStyle};
  max-width: 100%;
  border-radius: 0.5rem;
  background-color: aliceblue;
  color: black;
  margin-bottom: 2rem;
  overflow-x: auto;
`;

const StyledTable = styled.table`
  table-layout: fixed;
  width: max-content;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 0.8rem;

  th,
  td {
    padding: 0.75rem;
    text-align: left;
    border-bottom: 1px solid #ddd;
    white-space: wrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  td:empty {
    padding-right: 0;
    padding-left: 0;
  }

  td:has([data-collapsible]) {
    padding-right: 0;
  }

  td:has(.language-container) {
    white-space: normal;
  }

  td:has(.short-title-container) {
    white-space: nowrap;
    max-width: 0;
  }

  thead {
    position: sticky;
    top: 0;
    background-color: #f8f9fa;
    z-index: 1;
  }

  thead th {
    background-color: #f8f9fa;
  }

  tbody tr {
    transition: background-color 0.15s ease;
  }

  tbody tr:hover {
    background-color: #f0f0f0;
  }

  th {
    font-weight: bold;
    position: relative;
    padding-right: 1.5rem;

    &:hover {
      background-color: #eee;
    }

    .resizer {
      position: absolute;
      right: 0;
      top: 0;
      height: 100%;
      width: 4px;
      background: rgba(0, 0, 0, 0.05);
      cursor: col-resize;
      user-select: none;
      touch-action: none;

      &:hover {
        background: rgba(0, 0, 0, 0.2);
      }

      &.isResizing {
        background: rgba(0, 0, 0, 0.3);
        opacity: 1;
      }
    }
  }

  @media (max-width: 768px) {
    font-size: 0.8rem;

    th,
    td {
      padding: 0.5rem;
    }
  }
`;

const SortIndicator = styled.span`
  margin-left: 0.5rem;
`;

const ViewButton = styled.button`
  background-color: #f0f0f0;
  color: black;
  border: 1px solid #ccc;
  border-radius: 0.25rem;
  padding: 0.25rem 0.5rem;
  cursor: pointer;
  font-size: 0.85rem;

  &:hover {
    background-color: #e0e0e0;
  }
`;

const IconButton = styled.button`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  color: ${SEA_COLOR};

  &:hover {
    opacity: 0.85;
  }

  &:focus-visible {
    outline: 2px solid ${SEA_COLOR};
    outline-offset: 2px;
    border-radius: 2px;
  }
`;

const LanguageSpan = styled.span`
  display: inline-block;
  margin-right: 0.25rem;
  padding: 0.125rem 0.25rem;
  background-color: #e0e0e0;
  border-radius: 0.25rem;
  font-size: 0.8rem;
`;

const ExportButton = styled.button`
  background-color: ${SEA_COLOR};
  color: white;
  border: none;
  border-radius: 0.25rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  font-size: 0.9rem;

  &:hover {
    opacity: 0.9;
  }
`;

const ViewModeToggle = styled.div`
  display: flex;
  gap: 0.5rem;
  align-items: center;
`;

const ExpandIcon = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1rem;
  height: 1rem;
  cursor: pointer;
  color: ${SEA_COLOR};
`;

const ChildRow = styled.tr`
  background-color: #f8f9fa !important;

  &:hover {
    background-color: #e9ecef !important;
  }
`;

type ViewMode = "flat" | "reprint";

type ItemWithCluster = Item & {
  isClusterRoot?: boolean;
  clusterKey?: string;
  clusterMembers?: ItemWithCluster[];
  isReprintOf?: string;
};

const columnHelper = createColumnHelper<ItemWithCluster>();
const EMPTY_ITEMS: ItemWithCluster[] = [];

const SORT_TO_SERVER_FIELD: Record<string, string> = {
  year: "year",
  cities: "cities",
  languages: "languages",
  editors: "editor",
  title: "shortTitle",
  format: "format",
  volumesCount: "volumes",
  additionalContent: "additionalContent",
  type: "isElements",
  study_corpora: "corpus",
};

const toServerOrderBy = (sorting: SortingState): search_OrderByOption[] => {
  const mapped = sorting
    .map((rule) => ({
      field: SORT_TO_SERVER_FIELD[rule.id],
      descending: rule.desc,
    }))
    .filter((rule) => rule.field);
  if (mapped.length === 0) {
    return [
      { field: "year", descending: false },
      { field: "key", descending: false },
    ];
  }
  return [...mapped, { field: "key", descending: false }];
};

export function Catalogue() {
  const { filters } = useAppliedFilter();
  const { token } = useContext(AuthContext);
  const [sorting, setSorting] = useState<SortingState>([
    { id: "year", desc: false },
  ]);
  const [columnResizeMode] = useState<ColumnResizeMode>("onChange");
  const [showScrollTop, setShowScrollTop] = useState(false);
  const [selectedItem, setSelectedItem] = useState<Item | null>(null);
  const [isExporting, setIsExporting] = useState(false);
  const [viewMode, setViewMode] = useLocalStorage<ViewMode>(
    "catalogue-view-mode",
    "reprint",
  );
  const [expanded, setExpanded] = useState<ExpandedState>({});
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const orderBy = useMemo(() => toServerOrderBy(sorting), [sorting]);
  const { data: localFacsimiles = [] } = useQuery({
    queryKey: ["facsimiles", "download-available"],
    queryFn: () => FacsimilesService.getFacsimilies({}),
  });
  const downloadAvailableEditionKeys = useMemo(
    () =>
      new Set(
        localFacsimiles
          .filter(
            (facsimile) => facsimile.download_available && facsimile.edition_id,
          )
          .map((facsimile) => facsimile.edition_id!),
      ),
    [localFacsimiles],
  );

  const copyEditionKey = useCallback(async (editionKey: string) => {
    try {
      await navigator.clipboard.writeText(editionKey);
      setCopiedKey(editionKey);
      window.setTimeout(() => {
        setCopiedKey((current) => (current === editionKey ? null : current));
      }, 2000);
    } catch (err) {
      console.error("Failed to copy edition key:", err);
    }
  }, []);
  const openMainScan = useCallback(
    (editionKey: string) => {
      if (!token) {
        return;
      }
      void openAuthenticatedFacsimilePDF(editionKey, token).catch((error) => {
        console.error("Failed to open main scan:", error);
      });
    },
    [token],
  );
  const {
    items: filteredItems,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    fetchAllItemsForExport,
  } = useEditionsSearchInfinite({
    pageSize: 25,
    orderBy,
  });

  useAutoOpenEditionFromQuery(filteredItems, setSelectedItem);

  const scrollToTop = () => {
    document.getElementById(MAIN_CONTENT_ID)?.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  const handleExportCitations = async () => {
    setIsExporting(true);
    try {
      const itemsForExport = await fetchAllItemsForExport();
      const timestamp = new Date().toISOString().slice(0, 10);
      exportCitationsAsRTF(
        itemsForExport,
        `chicago_citations_${timestamp}.rtf`,
      );
    } finally {
      setIsExporting(false);
    }
  };

  useEffect(() => {
    const el = document.getElementById(MAIN_CONTENT_ID);
    if (!el) {
      return;
    }

    const onScroll = () => {
      const scrollTop = el.scrollTop;
      setShowScrollTop(scrollTop > 200);

      const remaining = el.scrollHeight - (el.scrollTop + el.clientHeight);
      if (remaining < 300 && hasNextPage && !isFetchingNextPage) {
        fetchNextPage();
      }
    };

    el.addEventListener("scroll", onScroll);
    return () => el.removeEventListener("scroll", onScroll);
  }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

  const processedItems = useMemo(() => {
    if (viewMode === "flat") {
      return filteredItems;
    }

    const normalizeKey = (key?: string | null) =>
      (key || "").trim().toLowerCase();
    const items = filteredItems ?? [];
    const itemMap = new Map(
      items.map((item) => [normalizeKey(item.key), item]),
    );
    const rootKeyByItemKey = new Map<string, string>();

    const sortByYearThenKey = (a: Item, b: Item) => {
      const yearA = a.year ? parseInt(a.year) : 9999;
      const yearB = b.year ? parseInt(b.year) : 9999;
      if (yearA !== yearB) {
        return yearA - yearB;
      }
      return a.key.localeCompare(b.key);
    };

    const resolveRootKey = (item: Item) => {
      const itemKey = normalizeKey(item.key);
      const cached = rootKeyByItemKey.get(itemKey);
      if (cached) {
        return cached;
      }

      let current: Item = item;
      const visited = new Set<string>();

      while (true) {
        const currentKey = normalizeKey(current.key);
        if (visited.has(currentKey)) {
          break;
        }
        visited.add(currentKey);

        const parentKey = normalizeKey(current.reprintOf);
        if (!parentKey) {
          break;
        }
        const parent = itemMap.get(parentKey);
        if (!parent) {
          break;
        }
        current = parent;
      }

      const rootKey = normalizeKey(current.key);
      visited.forEach((key) => rootKeyByItemKey.set(key, rootKey));
      return rootKey;
    };

    const groups = new Map<string, Item[]>();
    for (const item of items) {
      const rootKey = resolveRootKey(item);
      const existing = groups.get(rootKey);
      if (existing) {
        existing.push(item);
      } else {
        groups.set(rootKey, [item]);
      }
    }

    const result: ItemWithCluster[] = [];
    for (const groupItems of groups.values()) {
      if (groupItems.length <= 1) {
        result.push(groupItems[0]);
        continue;
      }

      const [root, ...members] = [...groupItems].sort(sortByYearThenKey);
      result.push({
        ...root,
        isClusterRoot: true,
        clusterKey: root.key,
        clusterMembers: members.map((member) => ({
          ...member,
          ...(member.reprintOf && itemMap.has(normalizeKey(member.reprintOf))
            ? { isReprintOf: member.reprintOf }
            : {}),
        })),
      });
    }

    return result;
  }, [filteredItems, viewMode]);

  const showOtherColumns =
    !filters?.type ||
    filters.type.length === 0 ||
    filters.type.some((item) => item.value === "Other");

  const showElementsColumns =
    !filters?.type ||
    filters.type.length === 0 ||
    filters.type.some((item) => item.value === "Elements");

  const tableItems = filteredItems ?? EMPTY_ITEMS;
  const hasManuscriptRows = tableItems.some(
    (item) => item.materialType === "Manuscript",
  );
  const hasPrintNonElementsRows = tableItems.some(
    (item) =>
      item.materialType !== "Manuscript" && item.type !== ItemTypes.elements,
  );
  const hasNonElementsRows = tableItems.some(
    (item) => item.type !== ItemTypes.elements,
  );
  const hasElementsRows = tableItems.some(
    (item) => item.type === ItemTypes.elements,
  );
  const hasPrintRows = tableItems.some(
    (item) => item.materialType !== "Manuscript",
  );

  const editorsHeader = hasPrintNonElementsRows
    ? "Author/Editor"
    : hasManuscriptRows
      ? "Composer"
      : "Editor";

  const columns = useMemo(
    () =>
      [
        viewMode === "reprint" &&
          columnHelper.accessor((row) => row, {
            id: "expand",
            header: "",
            enableSorting: false,
            cell: ({ row }) => {
              if (
                row.original.isClusterRoot &&
                row.original.clusterMembers?.length
              ) {
                return (
                  <ExpandIcon
                    onClick={row.getToggleExpandedHandler()}
                    data-collapsible
                    title={row.getIsExpanded() ? "Collapse" : "Expand"}
                  >
                    {row.getIsExpanded() ? (
                      <FaChevronDown />
                    ) : (
                      <FaChevronRight />
                    )}
                  </ExpandIcon>
                );
              }
              if (row.depth > 0) {
                return (
                  <div style={{ marginLeft: 20 }} title="Reprint of">
                    <PiArrowBendDownRightBold />
                  </div>
                );
              }
              return null;
            },
            size: 5,
          }),
        columnHelper.accessor((row) => row, {
          id: "actions",
          header: "",
          enableSorting: false,
          cell: (info) => (
            <Row gap={0.5} justifyStart>
              <ViewButton
                onClick={() => setSelectedItem(info.getValue())}
                title="Full View"
              >
                ⤢
              </ViewButton>
              {token && (
                <a
                  href={withAppBasePath(
                    `/item/edit?key=${info.row.original.key}`,
                  )}
                  target="_blank"
                  rel="noopener noreferrer"
                  title="Edit Item"
                >
                  <AiFillEdit style={{ color: SEA_COLOR, fontSize: "1rem" }} />
                </a>
              )}
              <IconButton
                type="button"
                title={
                  copiedKey === info.row.original.key
                    ? "Copied!"
                    : "Copy edition key"
                }
                aria-label="Copy edition key"
                onClick={() => void copyEditionKey(info.row.original.key)}
              >
                {copiedKey === info.row.original.key ? (
                  <FaCheck style={{ fontSize: "1rem" }} />
                ) : (
                  <AiOutlineCopy style={{ fontSize: "1rem" }} />
                )}
              </IconButton>
              {token &&
                downloadAvailableEditionKeys.has(info.row.original.key) && (
                  <IconButton
                    type="button"
                    onClick={() => openMainScan(info.row.original.key)}
                    title="View main scan"
                    aria-label="View main scan"
                  >
                    <FaFilePdf style={{ color: SEA_COLOR, fontSize: "1rem" }} />
                  </IconButton>
                )}
              {info.row.original.facsimiles.length > 0 && (
                <FacsimileLinks
                  facsimiles={info.row.original.facsimiles}
                  color={SEA_COLOR}
                />
              )}
              {info.row.original.diagramCropsAvailable && (
                <a
                  href={withAppBasePath(
                    `/diagrams?key=${info.row.original.key}`,
                  )}
                  target="_blank"
                  rel="noopener noreferrer"
                  title="View Diagrams"
                >
                  <SiMaterialdesign style={{ color: SEA_COLOR }} />
                </a>
              )}
            </Row>
          ),
          size: 108,
        }),
        columnHelper.accessor("year", {
          header: "Year",
          cell: (info) => formatDisplayYear(info.row.original),
          size: 10,
        }),
        columnHelper.accessor("cities", {
          header: "Cities",
          cell: (info) => joinArr(info.getValue()) || NO_CITY,
          size: 40,
        }),
        columnHelper.accessor("languages", {
          header: "Languages",
          cell: (info) => (
            <div className="language-container">
              {info.getValue().map((lang, i) => (
                <LanguageSpan key={i}>{lang}</LanguageSpan>
              ))}
            </div>
          ),
          size: 100,
        }),
        columnHelper.accessor("editors", {
          header: editorsHeader,
          cell: (info) => formatDisplayEditors(info.row.original),
          size: 160,
        }),
        showOtherColumns &&
          hasNonElementsRows &&
          columnHelper.accessor((row) => row, {
            id: "title",
            header: "Title",
            cell: (info) => {
              if (info.row.original.shortTitle) {
                const val = info.row.original.shortTitle;
                return (
                  <span className="short-title-container" title={val}>
                    {val}
                  </span>
                );
              }
              let val = info.row.original.title || "";
              val = upperFirst(
                val
                  .replaceAll(/-\s+/gi, "")
                  .replaceAll(/\[vol\. 1]:?\s*/gi, "")
                  .replaceAll(/\[general title page]:?\s*/gi, "")
                  .replace(/\.\s*$/g, "")
                  .trim()
                  .toLowerCase(),
              );
              if (val === "?") {
                val = "";
              }
              return (
                <span className="short-title-container" title={val}>
                  {val}
                </span>
              );
            },
            size: 160,
          }),
        hasManuscriptRows &&
          columnHelper.accessor("repository", {
            header: "Repository",
            cell: (info) => info.getValue(),
            size: 100,
          }),
        hasPrintRows &&
          columnHelper.accessor("format", {
            header: "Format",
            cell: (info) =>
              info.getValue() ? `${info.getValue()}º` : info.getValue(),
            size: 60,
          }),
        showElementsColumns &&
          hasElementsRows &&
          columnHelper.accessor("elementsBooks", {
            header: "Elements Books",
            enableSorting: false,
            cell: (info) => formatDisplayBooks(info.row.original),
            size: 105,
          }),
        showElementsColumns &&
          hasElementsRows &&
          columnHelper.accessor("additionalContent", {
            header: "Additional Content",
            cell: (info) => joinArr(info.getValue()),
            size: 140,
          }),
        inEuclidesMode() &&
          columnHelper.accessor("study_corpora", {
            header: () => <Row gap={0.5}>Study Corpus</Row>,
            cell: (info) =>
              joinArr(
                info
                  .getValue()
                  .sort()
                  .map((val) => STUDY_CORPUSES[val] || val),
              ),
            size: 120,
          }),
      ].filter(Boolean) as ColumnDef<ItemWithCluster>[],
    [
      viewMode,
      editorsHeader,
      showOtherColumns,
      hasNonElementsRows,
      hasManuscriptRows,
      hasPrintRows,
      showElementsColumns,
      hasElementsRows,
      token,
      copiedKey,
      copyEditionKey,
      downloadAvailableEditionKeys,
      openMainScan,
    ],
  );

  const table = useReactTable<ItemWithCluster>({
    data: processedItems ?? EMPTY_ITEMS,
    columns,
    state: {
      sorting,
      expanded,
    },
    onSortingChange: setSorting,
    onExpandedChange: setExpanded,
    getCoreRowModel: getCoreRowModel(),
    getExpandedRowModel: getExpandedRowModel(),
    getSubRows: (row) =>
      viewMode === "reprint" ? row.clusterMembers || [] : undefined,
    columnResizeMode,
    sortDescFirst: false,
  });

  return (
    <Container
      style={{ width: "100%", padding: "0 1rem", boxSizing: "border-box" }}
    >
      {showScrollTop && (
        <ScrollToTopButton onClick={scrollToTop} title="Scroll to top">
          ↑
        </ScrollToTopButton>
      )}

      <Stats />

      <Row gap={4}>
        <ViewModeToggle>
          <span>View Mode:</span>
          <Switch>
            <SwitchOption
              selected={viewMode === "flat"}
              onClick={() => setViewMode("flat")}
            >
              Flat
            </SwitchOption>
            <SwitchOption
              selected={viewMode === "reprint"}
              onClick={() => setViewMode("reprint")}
            >
              Group by publication
            </SwitchOption>
          </Switch>
        </ViewModeToggle>

        <ExportButton onClick={handleExportCitations} disabled={isExporting}>
          {isExporting
            ? "Exporting citations..."
            : "Export Citations (Chicago Style)"}
        </ExportButton>
      </Row>

      <TableContainer>
        <StyledTable>
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    colSpan={header.colSpan}
                    style={{
                      width: header.getSize(),
                    }}
                  >
                    {header.isPlaceholder ? null : (
                      <div
                        onClick={header.column.getToggleSortingHandler()}
                        style={
                          header.column.getCanSort()
                            ? {
                                display: "inline-flex",
                                cursor: "pointer",
                                userSelect: "none",
                              }
                            : {}
                        }
                      >
                        {flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                        {header.column.getCanSort() && (
                          <SortIndicator>
                            {{
                              asc: "↑",
                              desc: "↓",
                            }[header.column.getIsSorted() as string] || ""}
                          </SortIndicator>
                        )}
                      </div>
                    )}
                    <div
                      {...{
                        onMouseDown: header.getResizeHandler(),
                        onTouchStart: header.getResizeHandler(),
                        className: `resizer ${header.column.getIsResizing() ? "isResizing" : ""}`,
                      }}
                    />
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => {
              const isChildRow = row.depth > 0;
              const RowComponent = isChildRow ? ChildRow : "tr";

              return (
                <RowComponent key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <td
                      key={cell.id}
                      style={{
                        width: cell.column.getSize(),
                      }}
                    >
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext(),
                      )}
                    </td>
                  ))}
                </RowComponent>
              );
            })}
          </tbody>
        </StyledTable>
      </TableContainer>

      {selectedItem && (
        <ItemModal
          item={selectedItem}
          featuresById={{}}
          onClose={() => setSelectedItem(null)}
        />
      )}
    </Container>
  );
}
