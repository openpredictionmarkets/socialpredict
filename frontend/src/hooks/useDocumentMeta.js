import { useEffect } from 'react';

/**
 * useDocumentMeta — sets document.title and OpenGraph meta tags for a page.
 *
 * For a React SPA the tags are set client-side. Crawlers that execute JS
 * (Googlebot, many social-share bots) will pick them up. For bots that
 * don't execute JS, a prerender service can be placed in front of the app.
 *
 * On unmount the values are reset to the site defaults.
 *
 * @param {Object} meta
 * @param {string} meta.title        - Page/tab title
 * @param {string} [meta.description] - og:description content
 * @param {string} [meta.url]         - og:url content (defaults to location.href)
 */
export function useDocumentMeta({ title, description, url } = {}) {
  useEffect(() => {
    const defaultTitle = 'SocialPredict';
    const defaultDescription = 'Prediction markets for the social web';

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

    return () => {
      document.title = prevTitle;
      ogTitle.setAttribute('content', defaultTitle);
      ogDesc.setAttribute('content', defaultDescription);
      ogUrl.setAttribute('content', '');
      ogType.setAttribute('content', 'website');
    };
  }, [title, description, url]);
}

export default useDocumentMeta;
