import styled from "@emotion/styled";
import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import type { model_Edition } from "@hub-api";
import {
  applyReprints,
  searchEditionsPage,
  type ReprintRelationship,
} from "../api/editionApi";
import { SEA_COLOR } from "../utils/colors";
import { withAppBasePath } from "../utils/basePath";
import { Modal, ModalClose, ModalContent } from "./tps/modal/ModalComponents";

const Content = styled(ModalContent)`
  justify-content: flex-start;
  min-width: min(90vw, 1050px);
`;

const Header = styled.div`
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;

  h2 {
    margin: 0 0 0.25rem;
  }
  p {
    margin: 0;
    color: #52606d;
  }
`;

const Actions = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  margin: 0.75rem 0;
  flex-wrap: wrap;
`;

const Button = styled.button<{ primary?: boolean }>`
  border: 1px solid ${SEA_COLOR};
  background: ${({ primary }) => (primary ? SEA_COLOR : "white")};
  color: ${({ primary }) => (primary ? "white" : SEA_COLOR)};
  border-radius: 0.3rem;
  padding: 0.5rem 0.85rem;
  cursor: pointer;

  &:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }
`;

const TableWrap = styled.div`
  overflow: auto;
  max-height: min(65vh, 900px);
  border: 1px solid #d7e0e8;
  border-radius: 0.4rem;
`;

const Table = styled.table`
  width: 100%;
  border-collapse: collapse;
  font-size: 0.88rem;

  th,
  td {
    padding: 0.7rem;
    text-align: left;
    border-bottom: 1px solid #d7e0e8;
    vertical-align: top;
  }
  th {
    background: #f3f7fa;
    position: sticky;
    top: 0;
  }
  tbody tr:last-of-type td {
    border-bottom: 0;
  }
  a {
    color: ${SEA_COLOR};
    font-weight: 600;
  }
`;

const EditionMeta = styled.div`
  margin-top: 0.2rem;
  color: #52606d;
  font-size: 0.8rem;
`;

const ErrorText = styled.p`
  color: #a61b1b;
  margin: 0.5rem 0;
`;

type Props = {
  candidates: ReprintRelationship[];
  token: string;
  onClose: () => void;
  onApplied: (updated: number, skipped: number) => void;
};

const editionUrl = (key: string) =>
  withAppBasePath(`/item/edit?key=${encodeURIComponent(key)}`);

const editionLabel = (
  edition: model_Edition | undefined,
  fallbackKey: string,
) => edition?.shortTitle || edition?.title || fallbackKey;

const editionEditorNames = (edition: model_Edition | undefined) => {
  const editors = (edition as model_Edition & { editor?: string[] })?.editor;
  return editors?.filter(Boolean).join(", ") || "";
};

const editionMetaLine = (edition: model_Edition | undefined, key: string) =>
  [key, edition?.cities?.join(", "), edition?.year, editionEditorNames(edition)]
    .filter(Boolean)
    .join(" · ");

export function ReprintDetectionModal({
  candidates,
  token,
  onClose,
  onApplied,
}: Props) {
  const allKeys = useMemo(() => {
    const keys = new Set<string>();
    for (const candidate of candidates) {
      keys.add(candidate.editionKey);
      keys.add(candidate.reprintOf);
    }
    return Array.from(keys);
  }, [candidates]);
  const [selected, setSelected] = useState(() => new Set<string>());
  const [isApplying, setIsApplying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const editionsQuery = useQuery({
    queryKey: ["editions", "reprint-review", allKeys],
    queryFn: () =>
      searchEditionsPage({
        fields_filter: { key: allKeys },
        offset: 0,
        limit: allKeys.length,
      }),
    enabled: allKeys.length > 0,
  });
  const editionsByKey = useMemo(
    () =>
      new Map(
        (editionsQuery.data?.items || [])
          .filter((edition): edition is model_Edition & { key: string } =>
            Boolean(edition.key),
          )
          .map((edition) => [edition.key, edition]),
      ),
    [editionsQuery.data],
  );

  useEffect(() => {
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !isApplying) onClose();
    };
    window.addEventListener("keydown", closeOnEscape);
    return () => window.removeEventListener("keydown", closeOnEscape);
  }, [isApplying, onClose]);

  const toggle = (key: string) => {
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const confirm = async () => {
    setIsApplying(true);
    setError(null);
    try {
      const result = await applyReprints(
        token,
        candidates
          .filter((candidate) => selected.has(candidate.editionKey))
          .map(({ editionKey, reprintOf }) => ({ editionKey, reprintOf })),
      );
      onApplied(result.updated?.length ?? 0, result.skipped?.length ?? 0);
    } catch (cause) {
      setError(
        cause instanceof Error ? cause.message : "Could not update reprints.",
      );
      setIsApplying(false);
    }
  };

  return (
    <Modal
      onClick={isApplying ? undefined : onClose}
      role="dialog"
      aria-modal="true"
      aria-labelledby="reprint-dialog-title"
    >
      <Content hasImage={false} onClick={(event) => event.stopPropagation()}>
        <ModalClose title="Close" onClick={isApplying ? undefined : onClose}>
          ×
        </ModalClose>
        <Header>
          <div>
            <h2 id="reprint-dialog-title">Review suspected reprints</h2>
            <p>
              Changes will be applied only to the rows you select and confirm.
            </p>
          </div>
        </Header>

        {candidates.length === 0 ? (
          <p>No new suspected reprints were found.</p>
        ) : (
          <>
            <Actions>
              <div>
                <Button
                  type="button"
                  onClick={() =>
                    setSelected(
                      new Set(
                        candidates.map((candidate) => candidate.editionKey),
                      ),
                    )
                  }
                >
                  Select all
                </Button>{" "}
                <Button type="button" onClick={() => setSelected(new Set())}>
                  Clear all
                </Button>
              </div>
              <span>
                {selected.size} of {candidates.length} selected
              </span>
            </Actions>
            <TableWrap>
              <Table>
                <thead>
                  <tr>
                    <th aria-label="Select" />
                    <th>Suspected reprint</th>
                    <th>Original edition</th>
                  </tr>
                </thead>
                <tbody>
                  {candidates.map((candidate) => (
                    <tr key={candidate.editionKey}>
                      <td>
                        <input
                          type="checkbox"
                          checked={selected.has(candidate.editionKey)}
                          onChange={() => toggle(candidate.editionKey)}
                          aria-label={`Select ${candidate.editionKey}`}
                        />
                      </td>
                      <td>
                        <a
                          href={editionUrl(
                            editionsByKey.get(candidate.editionKey)?.key ||
                              candidate.editionKey,
                          )}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          {editionLabel(
                            editionsByKey.get(candidate.editionKey),
                            candidate.editionKey,
                          )}
                        </a>
                        <EditionMeta>
                          {editionMetaLine(
                            editionsByKey.get(candidate.editionKey),
                            candidate.editionKey,
                          )}
                        </EditionMeta>
                      </td>
                      <td>
                        <a
                          href={editionUrl(
                            editionsByKey.get(candidate.reprintOf)?.key ||
                              candidate.reprintOf,
                          )}
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          {editionLabel(
                            editionsByKey.get(candidate.reprintOf),
                            candidate.reprintOf,
                          )}
                        </a>
                        <EditionMeta>
                          {editionMetaLine(
                            editionsByKey.get(candidate.reprintOf),
                            candidate.reprintOf,
                          )}
                        </EditionMeta>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </TableWrap>
            {editionsQuery.isLoading && <p>Loading edition details…</p>}
            {editionsQuery.isError && (
              <ErrorText role="alert">
                Could not load edition details.
              </ErrorText>
            )}
          </>
        )}
        {error && <ErrorText role="alert">{error}</ErrorText>}
        <Actions>
          <Button type="button" onClick={onClose} disabled={isApplying}>
            Cancel
          </Button>
          {candidates.length > 0 && (
            <Button
              primary
              type="button"
              onClick={confirm}
              disabled={isApplying || selected.size === 0}
            >
              {isApplying
                ? "Updating…"
                : `Confirm ${selected.size} reprint${selected.size === 1 ? "" : "s"}`}
            </Button>
          )}
        </Actions>
      </Content>
    </Modal>
  );
}
