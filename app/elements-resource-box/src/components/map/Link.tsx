type LinkProps = {
  url: string;
  text: string;
};

export const Link = ({ url, text }: LinkProps) => (
  <a href={url} target="_blank" rel="noreferrer noopener" title={url}>
    {text}
  </a>
);
