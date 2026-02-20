import { useCallback, useState } from "react";

import { useEventListener, useIsomorphicLayoutEffect } from "usehooks-ts";

interface Size {
  width: number;
  height: number;
}

export function useElementSize<T extends HTMLElement = HTMLDivElement>() {
  const [ref, setRef] = useState<T | null>(null);
  const [size, setSize] = useState<Size>({
    width: 0,
    height: 0,
  });
  const [location, setLocation] = useState<[number, number]>([0, 0]);

  const handleSize = useCallback(() => {
    setSize({
      width: ref?.offsetWidth || 0,
      height: ref?.offsetHeight || 0,
    });
    setLocation([ref?.offsetLeft || 0, ref?.offsetTop || 0]);
  }, [ref?.offsetHeight, ref?.offsetLeft, ref?.offsetTop, ref?.offsetWidth]);

  useEventListener("resize", handleSize);

  useIsomorphicLayoutEffect(() => {
    handleSize();
  }, [ref?.offsetHeight, ref?.offsetLeft, ref?.offsetHeight, ref?.offsetWidth]);

  return { ref: setRef, size, location, refresh: handleSize };
}
