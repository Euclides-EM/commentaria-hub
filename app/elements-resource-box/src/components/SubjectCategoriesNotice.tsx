import styled from "@emotion/styled";
import { buildCommentariaHubFeatureResultsUrl } from "../utils/commentariaHub";
import { SEA_COLOR } from "../utils/colors";

const Notice = styled.div`
  padding: 0.65rem 0.75rem;
  border-left: 3px solid #d59b24;
  background: #fff8e6;
  color: #5b4618;
  font-size: 0.82rem;
  line-height: 1.35;

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
      Some of the subject categories may have been calculated with the
      assistance of an LLM and may contain errors.{" "}
      {featureResultsUrl && (
        <>
          <a href={featureResultsUrl} target="_blank" rel="noopener noreferrer">
            Review the Feature Results
          </a>
          .
        </>
      )}
    </Notice>
  );
};
