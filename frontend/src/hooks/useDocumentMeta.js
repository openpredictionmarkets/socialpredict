import { useEffect } from 'react';

/**
 * useDocumentMeta — sets document.title and OpenGraph meta tags for a page.
 *
 * For crawler compatibility, public market pages should also be served by the
 * backend share shell. This hook keeps browser navigation metadata in sync
 * after the SPA hydrates.
 *
 * On unmount the values are reset to the site defaults.
 *
 * @param {Object} meta
 * @param {string} meta.title        - Page/tab title
 * @param {string} [meta.description] - og:description content
 * @param {string} [meta.url]         - og:url content (defaults to location.href)
 * @param {string} [meta.image]       - og:image and twitter:image content
 */
export function useDocumentMeta({ title, description, url, image } = {}) {
  useEffect(() => {
    const defaultTitle = 'SocialPredict';
    const defaultDescription = 'Prediction markets for the social web';
    const defaultImage = typeof window !== 'undefined' ? `${window.location.origin}/logo512.png` : '';

    const prevTitle = document.title;

    function setMeta(property, content) {
      let el = document.querySelector(`meta[property="${property}"]`);
      if (!el) {
        el = document.createElement('meta');
        el.setAttribute('property', property);
        document.head.appendChild(el);
      }
      el.setAttribute('content', content);
      return el;
    }

    if (title) document.title = title;
    const ogTitle = setMeta('og:title', title || defaultTitle);
    const ogDesc = setMeta('og:description', description || defaultDescription);
    const ogUrl = setMeta('og:url', url || (typeof window !== 'undefined' ? window.location.href : ''));
    const ogType = setMeta('og:type', 'website');
    const ogImage = setMeta('og:image', image || defaultImage);
    const ogSiteName = setMeta('og:site_name', 'SocialPredict');

    function setNameMeta(name, content) {
      let el = document.querySelector(`meta[name="${name}"]`);
      if (!el) {
        el = document.createElement('meta');
        el.setAttribute('name', name);
        document.head.appendChild(el);
      }
      el.setAttribute('content', content);
      return el;
    }

    const twitterCard = setNameMeta('twitter:card', 'summary_large_image');
    const twitterTitle = setNameMeta('twitter:title', title || defaultTitle);
    const twitterDesc = setNameMeta('twitter:description', description || defaultDescription);
    const twitterImage = setNameMeta('twitter:image', image || defaultImage);

    return () => {
      document.title = prevTitle;
      ogTitle.setAttribute('content', defaultTitle);
      ogDesc.setAttribute('content', defaultDescription);
      ogUrl.setAttribute('content', '');
      ogType.setAttribute('content', 'website');
      ogImage.setAttribute('content', defaultImage);
      ogSiteName.setAttribute('content', 'SocialPredict');
      twitterCard.setAttribute('content', 'summary_large_image');
      twitterTitle.setAttribute('content', defaultTitle);
      twitterDesc.setAttribute('content', defaultDescription);
      twitterImage.setAttribute('content', defaultImage);
    };
  }, [title, description, url, image]);
}

export default useDocumentMeta;
