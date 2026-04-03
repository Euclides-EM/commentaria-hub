import React, { createContext, ReactNode, useContext, useMemo } from "react";
import { useLocalStorage } from "usehooks-ts";

type FilterEditContextType = {
  filterOpen: boolean;
  setFilterOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

const FilterEditContext = createContext<FilterEditContextType | undefined>(
  undefined,
);

export const useEditFilter = () => {
  const context = useContext(FilterEditContext);
  if (!context) {
    throw new Error("useEditFilter must be used within a FilterEditProvider");
  }
  return context;
};

export const FilterEditProvider = ({ children }: { children: ReactNode }) => {
  const [filterOpen, setFilterOpen] = useLocalStorage<boolean>(
    "filters-open",
    false,
  );

  const value = useMemo(
    () => ({ filterOpen, setFilterOpen }),
    [filterOpen, setFilterOpen],
  );

  return (
    <FilterEditContext.Provider value={value}>
      {children}
    </FilterEditContext.Provider>
  );
};
