import styled from "@emotion/styled";
import { MARKER_1, MARKER_2, MARKER_5, SEA_COLOR } from "../../utils/colors";

export const PageContainer = styled.div`
  padding: 1.5rem;
  width: 100%;
  margin: 1.5rem auto;
  max-width: min(1200px, 100vw - 4rem);
  min-height: calc(100vh - 120px);
  color: black;
  background-color: aliceblue;
  border-radius: 1rem;
  box-sizing: border-box;
`;

export const PageHeader = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
`;

export const Title = styled.h1`
  margin: 0;
  font-size: 2rem;
  color: black;
`;

export const Card = styled.div`
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 0.75rem;
  padding: 1.25rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
`;

export const Section = styled.div`
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-bottom: 2rem;
`;

export const FormGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 1rem;
`;

export const CreateFeatureGrid = styled.div`
  display: grid;
  grid-template-columns: minmax(320px, 1fr) auto;
  gap: 1rem;
  align-items: start;

  @media (max-width: 900px) {
    grid-template-columns: 1fr;
  }
`;

export const CreateFeatureMainColumn = styled.div`
  display: flex;
  flex-direction: column;
  gap: 1rem;
`;

export const Field = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
`;

export const Label = styled.label`
  font-size: 0.85rem;
  color: #666;
  font-weight: 600;
`;

export const Input = styled.input`
  padding: 0.55rem 0.65rem;
  border: 1px solid #ccc;
  border-radius: 0.4rem;
  background-color: white;
  color: black;
  font-size: 0.9rem;

  &:focus {
    outline: none;
    border-color: ${SEA_COLOR};
    background-color: white;
  }
`;

export const SearchInput = styled(Input)`
  width: 100%;
  max-width: 520px;
`;

export const SearchRow = styled.div`
  display: flex;
  align-items: center;
  gap: 1rem;
`;

export const TextArea = styled.textarea`
  padding: 0.55rem 0.65rem;
  border: 1px solid #ccc;
  border-radius: 0.4rem;
  background-color: white;
  color: black;
  font-size: 0.9rem;
  resize: vertical;
  min-height: 80px;

  &:focus {
    outline: none;
    border-color: ${SEA_COLOR};
    background-color: white;
  }
`;

export const Select = styled.select`
  padding: 0.55rem 0.65rem;
  border: 1px solid #ccc;
  border-radius: 0.4rem;
  background-color: white;
  color: black;
  font-size: 0.9rem;

  &:focus {
    outline: none;
    border-color: ${SEA_COLOR};
    background-color: white;
  }
`;

export const ButtonRow = styled.div`
  display: flex;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
`;

export const CollapseButton = styled.button`
  border: none;
  background: transparent;
  color: ${SEA_COLOR};
  border-radius: 999px;
  width: 1.5rem;
  height: 1.5rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 1rem;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: #f0f0f0;
  }
`;

export const Form = styled.form`
  display: flex;
  flex-direction: column;
  gap: 1rem;
`;

export const Button = styled.button<{
  variant?: "primary" | "danger" | "ghost" | "outline";
}>`
  padding: 0.5rem 1rem;
  min-width: 6.5rem;
  text-align: center;
  border-radius: 0.4rem;
  border: 1px solid
    ${({ variant }) => (variant === "outline" ? SEA_COLOR : "transparent")};
  background-color: ${({ variant }) =>
    variant === "danger"
      ? MARKER_2
      : variant === "outline"
        ? "transparent"
        : variant === "ghost"
          ? "transparent"
          : SEA_COLOR};
  color: ${({ variant }) =>
    variant === "ghost" || variant === "outline" ? SEA_COLOR : "white"};
  ${({ variant }) =>
    variant === "ghost" &&
    `
      color: ${SEA_COLOR};
      background-color: #f0f0f0;
    `};
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  transition: opacity 0.2s ease;

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  &:hover:not(:disabled) {
    opacity: 0.85;
  }
`;

export const FeatureList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
`;

export const FeatureHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 1rem;
`;

export const FeatureTitle = styled.div<{ clickable?: boolean }>`
  font-size: 1.1rem;
  font-weight: 700;
  color: black;
  ${({ clickable }) => clickable && "cursor: pointer;"}
`;

export const FeatureDescription = styled.div`
  font-size: 0.9rem;
  color: #666;
  line-height: 1.4;
  white-space: pre-wrap;
`;

export const FeatureMeta = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
`;

export const FeatureActions = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
`;

export const FeatureTitleRow = styled.div`
  display: flex;
  align-items: center;
  gap: 0.6rem;
`;

export const ColorSwatch = styled.span<{ color?: string }>`
  width: 0.9rem;
  height: 0.9rem;
  border-radius: 0.3rem;
  border: 1px solid #ccc;
  background-color: ${({ color }) => color || "#f2f2f2"};
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.4);
  flex-shrink: 0;
`;

export const ColorPickerRow = styled.div`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
`;

export const ColorPickerPanel = styled.div`
  .react-colorful {
    width: 220px;
    height: 170px;
  }

  .react-colorful__saturation,
  .react-colorful__hue,
  .react-colorful__alpha {
    border-radius: 0.5rem;
  }

  .react-colorful__saturation-pointer,
  .react-colorful__hue-pointer,
  .react-colorful__alpha-pointer {
    width: 14px;
    height: 14px;
  }
`;

export const ColorHexInput = styled.input`
  padding: 0.55rem 0.65rem;
  border: 1px solid #ccc;
  border-radius: 0.4rem;
  background-color: white;
  color: black;
  font-size: 0.9rem;
  width: 120px;
  font-family:
    "SFMono-Regular", Menlo, Monaco, Consolas, "Liberation Mono", "Courier New",
    monospace;

  &:focus {
    outline: none;
    border-color: ${SEA_COLOR};
    background-color: white;
  }
`;

export const IconButton = styled.span`
  border: none;
  background: transparent;
  color: ${SEA_COLOR};
  border-radius: 999px;
  width: 1.6rem;
  height: 1.6rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  transition: background-color 0.2s ease;

  &:hover {
    background-color: #f0f0f0;
  }
`;

export const Tag = styled.span`
  background-color: ${MARKER_5};
  color: ${MARKER_1};
  border-radius: 999px;
  padding: 0.2rem 0.65rem;
  font-size: 0.75rem;
  font-weight: 600;
`;

export const RevisionList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 0.75rem;
`;

export const RevisionHeader = styled.button`
  border: none;
  background: transparent;
  padding: 0;
  display: flex;
  align-items: center;
  gap: 0;
  margin-left: -0.5rem;
  cursor: pointer;
`;

export const RevisionLabel = styled.span`
  font-size: 0.85rem;
  color: #666;
`;

export const RevisionCard = styled.div`
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 0.5rem;
  padding: 0.75rem 0.9rem;
  font-size: 0.85rem;
  display: grid;
  gap: 0.4rem;
  position: relative;
  white-space: pre-wrap;
`;

export const LatestTag = styled(Tag)`
  position: absolute;
  top: 0.6rem;
  right: 0.75rem;
`;
export const InlineValue = styled.span`
  color: black;
  font-weight: 600;
`;

export const SmallText = styled.div`
  font-size: 0.8rem;
  color: #666;
`;

export const EmptyState = styled.div`
  font-size: 0.9rem;
  color: #666;
`;

export const ErrorText = styled.div`
  color: ${MARKER_2};
  font-size: 0.9rem;
`;

export const TabBar = styled.div`
  display: flex;
  flex-direction: row;
  margin-bottom: 1.5rem;
`;

export const TabButton = styled.button<{ active: boolean }>`
  cursor: pointer;
  padding: 0.5rem 1.25rem;
  font-size: 0.9rem;
  font-weight: ${({ active }) => (active ? 600 : 400)};
  border: 1px solid ${({ active }) => (active ? SEA_COLOR : "#ddd")};
  background-color: ${({ active }) => (active ? SEA_COLOR : "transparent")};
  color: ${({ active }) => (active ? "white" : "black")};
  transition:
    background-color 0.2s ease,
    color 0.2s ease;

  &:first-of-type {
    border-radius: 0.4rem 0 0 0.4rem;
  }

  &:last-of-type {
    border-radius: 0 0.4rem 0.4rem 0;
  }

  &:hover {
    opacity: 0.85;
  }
`;

export const CheckboxRow = styled.label`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.9rem;
  padding: 0.3rem 0;
`;

export const CheckboxList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #ccc;
  border-radius: 0.4rem;
  padding: 0.5rem;
  background-color: white;
`;

export const ItemRow = styled.label<{ interactive?: boolean }>`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: ${({ interactive = false }) => (interactive ? "pointer" : "default")};
  font-size: 0.85rem;
  padding: 0.35rem 0.25rem;
  border-bottom: 1px solid #f0f0f0;
  min-width: 0;

  &:last-of-type {
    border-bottom: none;
  }
`;

export const ItemDetails = styled.span`
  color: #333;
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
`;

export const NoRevisionText = styled.span`
  font-size: 0.8rem;
  color: #666;
  margin-left: auto;
`;

export const ExecutionEditionsToggle = styled.div`
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  font-size: 0.8rem;
  color: #666;
`;

export const ExecutionEditionsToggleContent = styled.span`
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
`;

export const StatusTag = styled(Tag)<{ status?: string }>`
  background-color: ${({ status }) =>
    status === "success"
      ? "#d4edda"
      : status === "failed"
        ? "#f8d7da"
        : status === "in_progress"
          ? "#fff3cd"
          : status === "canceling"
            ? "#fff3cd"
            : status === "canceled"
              ? "#e2e3e5"
              : MARKER_5};
  color: ${({ status }) =>
    status === "success"
      ? "#155724"
      : status === "failed"
        ? "#721c24"
        : status === "in_progress"
          ? "#856404"
          : status === "canceling"
            ? "#856404"
            : status === "canceled"
              ? "#383d41"
              : MARKER_1};
`;

export const FeatureTokenList = styled.div`
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
  margin-top: 0.75rem;
`;

export const FeatureToken = styled.span`
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background-color: #f3f6fa;
  border: 1px solid #d6dde5;
  border-radius: 999px;
  padding: 0.2rem 0.55rem;
  color: #2a3642;
  font-size: 0.75rem;
  line-height: 1.2;
  max-width: 100%;
`;

export const FeatureTokenColor = styled.span<{ color?: string }>`
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 999px;
  border: 1px solid #c6cbd2;
  background-color: ${({ color }) => color || "#d7dce2"};
  flex-shrink: 0;
`;
