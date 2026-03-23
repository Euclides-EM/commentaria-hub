import { sortBy } from "lodash";
import { useMemo } from "react";
import styled from "@emotion/styled";
import { PANE_COLOR, TRANSPARENT_BLACK } from "../../utils/colors";
import TopDeco from "./svg/deco3.svg?react";
import BottomDeco from "./svg/deco2.svg?react";
import { Item } from "../../types";
import { NO_CITY } from "../../constants";
import pluralize from "pluralize";

type CityDetailsProps = {
  city: string;
  data: Item[];
  selectedRecordId: string | undefined;
  setSelectedRecordKey: (selected: string) => void;
};

const Title = styled.div`
  background-color: ${TRANSPARENT_BLACK};
  color: ${PANE_COLOR};
  width: calc(100% - 1rem);
  padding: 0.5rem;
  border-radius: 0.5rem;
  text-align: center;
`;

const Separator = styled.div`
  height: 1px;
  width: 100%;
  background-color: black;
  margin-bottom: 1rem;
`;

const Subtitle = styled.div`
  font-size: 0.8rem;
`;

const RowTitle = styled.div`
  display: flex;
  flex-direction: column;
  align-items: start;
  gap: 0.5rem;
  cursor: pointer;
  width: fit-content;
  border-radius: 0.5rem;
  padding: 0.5rem;

  &:hover {
    background-color: ${TRANSPARENT_BLACK};
    color: ${PANE_COLOR};
  }
`;

const StyledTopDeco = styled(TopDeco)`
  transform: rotate(180deg);
  fill: rgba(0, 0, 0, 0.6);
  margin-top: -0.8rem;
`;

const StyledBottomDeco = styled(BottomDeco)`
  fill: rgba(0, 0, 0, 0.6);
  height: 400px;
  width: 100%;
`;

export const CityName = (city: string) =>
  city === NO_CITY ? "Uncategorized" : city;

export const CityDetails = ({
  city,
  data,
  setSelectedRecordKey,
}: CityDetailsProps) => {
  const sortedData = useMemo(() => sortBy(data, "year"), [data]);
  return (
    <>
      <Title className="gothic">{CityName(city)}</Title>
      <StyledTopDeco height={64} />
      <Subtitle>
        {data?.length || 0} {pluralize("record", data?.length || 0)}
      </Subtitle>
      {sortedData.map((item) => (
        <div key={item.key}>
          <Separator />
          <RowTitle onClick={() => setSelectedRecordKey(item.key)}>
            <div className="gothic" title="Year">
              {item.year || item.year || "Unknown"}
            </div>
            {item.editors.map((editor) => (
              <div title="Editor" key={editor}>
                <span className="gothic">{editor[0]}</span>
                {editor.substring(1)}
              </div>
            ))}
          </RowTitle>
        </div>
      ))}
      <div>
        <StyledBottomDeco viewBox="0 0 80 360" />
      </div>
    </>
  );
};
