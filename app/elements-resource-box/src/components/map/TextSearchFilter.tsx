import React from "react";
import styled from "@emotion/styled";
import { Item } from "../../types";

const Container = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
`;

const SearchInput = styled.input`
  padding: 0.5rem;
  border: 1px solid #ccc;
  background: white;
  color: black;
  border-radius: 4px;
  font-size: 0.9rem;
  width: 100%;
  box-sizing: border-box;

  &:focus {
    outline: none;
    border-color: #666;
  }
`;

const FieldSelector = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding-left: 2rem;
`;

const FilterTitle = styled.div`
  font-size: 1rem;
  margin-bottom: 0.5rem;
  color: black;
`;

const Label = styled.label`
  font-size: 0.9rem;
  margin-bottom: 0.25rem;
`;

const CheckboxGroup = styled.div`
  display: flex;
  flex-direction: row;
  gap: 1.5rem;
`;

const CheckboxLabel = styled.label`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  cursor: pointer;

  input[type="checkbox"] {
    margin: 0;
  }
`;

type TextSearchFilterProps = {
  textSearch: string;
  setTextSearch: (value: string) => void;
  textSearchFields: (keyof Item)[];
  setTextSearchFields: (fields: (keyof Item)[]) => void;
};

const FIELD_OPTIONS: { value: keyof Item; label: string }[] = [
  { value: "shortTitle", label: "Short Title" },
  { value: "title", label: "Title" },
  { value: "titleEn", label: "Title (English)" },
];

export const TextSearchFilter: React.FC<TextSearchFilterProps> = ({
  textSearch,
  setTextSearch,
  textSearchFields,
  setTextSearchFields,
}) => {
  const handleFieldChange = (field: keyof Item, checked: boolean) => {
    if (checked) {
      setTextSearchFields([...textSearchFields, field]);
    } else {
      setTextSearchFields(textSearchFields.filter((f) => f !== field));
    }
  };

  return (
    <Container>
      <div>
        <FilterTitle>
          <span className="gothic">T</span>ext Search
        </FilterTitle>
        <SearchInput
          type="text"
          placeholder="Search in titles..."
          value={textSearch}
          onChange={(e) => setTextSearch(e.target.value)}
        />
      </div>
      <FieldSelector>
        <Label>Search in:</Label>
        <CheckboxGroup>
          {FIELD_OPTIONS.map((option) => (
            <CheckboxLabel key={option.value}>
              <input
                type="checkbox"
                checked={textSearchFields.includes(option.value)}
                onChange={(e) =>
                  handleFieldChange(option.value, e.target.checked)
                }
              />
              {option.label}
            </CheckboxLabel>
          ))}
        </CheckboxGroup>
      </FieldSelector>
    </Container>
  );
};
