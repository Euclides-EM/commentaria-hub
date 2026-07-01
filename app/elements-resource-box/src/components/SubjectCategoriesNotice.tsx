import styled from "@emotion/styled";
import { IoWarning } from "react-icons/io5";
import { buildCommentariaHubFeatureResultsUrl } from "../utils/commentariaHub";
import { SEA_COLOR } from "../utils/colors";

const Notice = styled.div`
  display: flex;
  align-items: baseline;
  gap: 0.35rem;
  opacity: 0.8;
  font-size: 0.82rem;
  line-height: 1.35;

  svg {
    flex: 0 0 auto;
  }

  a {
    color: ${SEA_COLOR};
    font-weight: 600;
  }
`;

type Props = {
  className?: string;
  editionKey?: string | null;
};

export const SubjectCategoriesNotice = ({ className, editionKey }: Props) => {
  const featureResultsUrl = buildCommentariaHubFeatureResultsUrl({
    featureId: "m_classifier",
    editionKey,
  });

  return (
    <Notice className={className}>
      <IoWarning />
      <span>
        <strong>Experimental</strong> Subject categories were generated with
        LLM assistance and are still being refined. They may contain errors, use
        with care.{" "}
        {featureResultsUrl && (
          <a href={featureResultsUrl} target="_blank" rel="noopener noreferrer">
            Review Feature Results
          </a>
        )}
        {featureResultsUrl && "."}
      </span>
    </Notice>
  );
};
