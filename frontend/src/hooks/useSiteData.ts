import { useEffect, useState } from 'react';
import { getSiteData } from '../services/siteApi';
import type { SiteData } from '../types/site';

interface SiteDataState {
  data: SiteData | null;
  error: Error | null;
  isLoading: boolean;
}

export function useSiteData(): SiteDataState {
  const [state, setState] = useState<SiteDataState>({
    data: null,
    error: null,
    isLoading: true,
  });

  useEffect(() => {
    let ignore = false;

    getSiteData()
      .then((data) => {
        if (!ignore) setState({ data, error: null, isLoading: false });
      })
      .catch((error: Error) => {
        if (!ignore) setState({ data: null, error, isLoading: false });
      });

    return () => {
      ignore = true;
    };
  }, []);

  return state;
}
