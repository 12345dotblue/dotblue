import React, { useState, useRef, useCallback } from 'react';
import { Select } from 'antd';
import axios from 'axios';
import { BACKEND_URL } from '../config';

interface EnterpriseOption {
  id: string;
  name: string;
  slug?: string;
}

interface EnterpriseSearchSelectProps {
  value?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  style?: React.CSSProperties;
  getHeaders?: () => Record<string, string>;
}

interface SearchResponse {
  items: EnterpriseOption[];
  total: number;
}

export default function EnterpriseSearchSelect({
  value,
  onChange,
  placeholder,
  style,
  getHeaders,
}: EnterpriseSearchSelectProps) {
  const [options, setOptions] = useState<{ label: string; value: string }[]>([]);
  const [loading, setLoading] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);

  const doSearch = useCallback(async (keyword: string) => {
    setLoading(true);
    try {
      const headers = getHeaders ? getHeaders() : {};
      const res = await axios.get<SearchResponse>(
        `${BACKEND_URL}/api/admin/platform/enterprises/search`,
        { headers, params: { keyword, page: 1, pageSize: 20 } },
      );
      const items = res.data?.items || [];
      setOptions(items.map((item) => ({ label: `${item.name} (${item.id})`, value: item.id })));
    } catch {
      setOptions([]);
    } finally {
      setLoading(false);
    }
  }, [getHeaders]);

  const handleSearch = (query: string) => {
    if (timerRef.current) clearTimeout(timerRef.current);
    if (!query.trim()) {
      setOptions([]);
      return;
    }
    timerRef.current = setTimeout(() => doSearch(query.trim()), 300);
  };

  return (
    <Select
      showSearch
      value={value}
      onChange={onChange}
      onSearch={handleSearch}
      filterOption={false}
      loading={loading}
      options={options}
      placeholder={placeholder}
      style={style}
      notFoundContent={loading ? undefined : null}
    />
  );
}
