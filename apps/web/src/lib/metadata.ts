interface MetadataProps {
  title: string;
  description: string;
  image?: string;
  isHome?: boolean;
  isNoIndex?: boolean;
}

export function createMetadata({
  title,
  description,
  image,
  isHome = false,
  isNoIndex = false,
}: MetadataProps) {
  const fullTitle = isHome ? title : `${title} | Code Self Study`;

  return [
    {
      title: fullTitle,
    },
    {
      name: "description",
      content: description,
    },
    {
      name: "robots",
      content: isNoIndex ? "noindex" : "max-image-preview:large",
    },
    {
      property: "og:title",
      content: title,
    },
    {
      property: "og:description",
      content: description,
    },
    {
      property: "og:type",
      content: "website",
    },
    ...(image
      ? [
          {
            property: "og:image",
            content: image,
          },
        ]
      : []),
  ];
}
