import { createContext } from "react";

type Auth = {
  token: string | null;
  setToken: (token: string | null) => void;
};

export const AuthContext = createContext<Auth>({
  token: null,
  setToken: () => {},
});
