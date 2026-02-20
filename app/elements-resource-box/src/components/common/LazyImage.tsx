import { useState, useRef, useEffect } from "react";
import styled from "@emotion/styled";

const ImageContainer = styled.div<{ width?: string; height?: string }>`
  width: ${({ width }) => width || "100%"};
  height: ${({ height }) => height || "12rem"};
  background-color: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
`;

const Image = styled.img`
  width: 100%;
  height: 100%;
  object-fit: contain;
  transition: opacity 0.3s;
`;

const Placeholder = styled.div`
  width: 100%;
  height: 100%;
  background-color: #f3f4f6;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  font-size: 0.875rem;
`;

interface LazyImageProps {
  src: string;
  alt: string;
  width?: string;
  height?: string;
  placeholder?: string;
  className?: string;
  onClick?: () => void;
  priority?: boolean;
}

export const LazyImage = ({
  src,
  alt,
  width,
  height,
  placeholder = "Loading...",
  className,
  onClick,
  priority = false,
}: LazyImageProps) => {
  const [isLoaded, setIsLoaded] = useState(false);
  const [isInView, setIsInView] = useState(priority);
  const [hasError, setHasError] = useState(false);
  const imgRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (priority) {
      const img = document.createElement("img");
      img.src = src;
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;
        if (entry.isIntersecting) {
          setIsInView(true);
          observer.disconnect();
        }
      },
      {
        threshold: 0.1,
        rootMargin: "200px",
      },
    );

    if (imgRef.current) {
      observer.observe(imgRef.current);
    }

    return () => observer.disconnect();
  }, [src, priority]);

  const handleLoad = () => {
    setIsLoaded(true);
  };

  const handleError = () => {
    setHasError(true);
    setIsLoaded(true);
  };

  return (
    <ImageContainer
      ref={imgRef}
      width={width}
      height={height}
      className={className}
      onClick={onClick}
    >
      {!isInView ? (
        <Placeholder>{placeholder}</Placeholder>
      ) : hasError ? (
        <Placeholder>Failed to load image</Placeholder>
      ) : (
        <>
          {!isLoaded && <Placeholder>{placeholder}</Placeholder>}
          <Image
            src={src}
            alt={alt}
            onLoad={handleLoad}
            onError={handleError}
            style={{ opacity: isLoaded ? 1 : 0 }}
            loading={priority ? "eager" : "lazy"}
          />
        </>
      )}
    </ImageContainer>
  );
};
