import styled from "@emotion/styled";
import { Row, Text } from "../../components/common";
import { GroupByOption } from "./useTrendsData";
import Select from "react-select";
import React from "react";

type TrendsControlsProps = {
  groupByOptions: GroupByOption[];
  groupBy: GroupByOption;
  handleGroupByChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  windowSize: number;
  setWindowSize: (size: number) => void;
};

const StyledSelect = styled(Select)<{ width: number }>`
  color: black;
  .select__control {
    min-width: ${(props) => props.width}rem;
  }
`;

export function TrendsControls({
  groupByOptions,
  groupBy,
  handleGroupByChange,
  windowSize,
  setWindowSize,
}: TrendsControlsProps) {
  return (
    <Row>
      <Text size={1.2}>Group by:</Text>
      <StyledSelect
        width={20}
        name="Group by"
        value={{
          value: groupBy.key,
          label:
            groupByOptions.find((opt) => opt.key === groupBy.key)?.label ||
            groupBy.key,
        }}
        options={groupByOptions.map((option) => ({
          value: option.key,
          label: option.label,
        }))}
        className="basic-select"
        classNamePrefix="select"
        onChange={(selected) => {
          if (selected) {
            const selectedValue = (selected as { value: string | null }).value;
            const mockEvent = {
              target: { value: selectedValue },
            } as React.ChangeEvent<HTMLSelectElement>;
            handleGroupByChange(mockEvent);
          }
        }}
        placeholder="Select Group by"
      />

      <Text size={1.2}>Year range:</Text>
      <StyledSelect
        width={10}
        name="Year range"
        value={{
          value: windowSize.toString(),
          label: `${windowSize} year${windowSize > 1 ? "s" : ""}`,
        }}
        options={[1, 5, 10, 20, 50].map((size) => ({
          value: size.toString(),
          label: `${size} year${size > 1 ? "s" : ""}`,
        }))}
        className="basic-select"
        classNamePrefix="select"
        onChange={(selected) => {
          if (selected) {
            const selectedValue = (selected as { value: string | null }).value;
            setWindowSize(Number(selectedValue));
          }
        }}
        placeholder="Select Year range"
      />
    </Row>
  );
}
