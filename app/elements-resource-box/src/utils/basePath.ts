const normalizeBasePath = (value?: string): string => {
  const raw = (value ?? "").trim();
  if (!raw || raw === "/") {
    return "/";
  }
  const withLeadingSlash = raw.startsWith("/") ? raw : `/${raw}`;
  return withLeadingSlash.endsWith("/")
    ? withLeadingSlash
    : `${withLeadingSlash}/`;
};

const getAppBasePath = (): string =>
  normalizeBasePath(import.meta.env.BASE_URL ?? "/");

export const getRouterBasename = (): string => {
  const base = getAppBasePath();
  return base === "/" ? "/" : base.slice(0, -1);
};

const stripBasePath = (path: string): string => {
  const base = getAppBasePath();
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  if (base === "/") {
    return normalizedPath;
  }

  const baseNoTrailing = base.slice(0, -1);
  if (normalizedPath === baseNoTrailing) {
    return "/";
  }

  if (normalizedPath.startsWith(base)) {
    const remainder = normalizedPath.slice(base.length - 1);
    return remainder.startsWith("/") ? remainder : `/${remainder}`;
  }

  return normalizedPath;
};

export const withAppBasePath = (path: string): string => {
  const base = getAppBasePath();
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  if (base === "/") {
    return normalizedPath;
  }

  const baseNoTrailing = base.slice(0, -1);
  if (
    normalizedPath === baseNoTrailing ||
    normalizedPath.startsWith(base) ||
    normalizedPath.startsWith(`${baseNoTrailing}/`)
  ) {
    return normalizedPath;
  }

  return `${baseNoTrailing}${normalizedPath}`;
};

export const getAppPathname = (): string =>
  stripBasePath(window.location.pathname);

export const buildAppUrl = (pathname: string, search = ""): string =>
  `${withAppBasePath(pathname)}${search}`;
