import { request } from '../shared/http'
import type { DataRujukanInternal, InputRujukan, ReferensiRujukan, RujukanInternal } from '../../types/rujukanInternal'

function path(ranap: boolean) {
  return ranap ? '/api/rujukan-internal-ranap' : '/api/rujukan-internal-poli'
}

export function dataRujukanInternal(token: string, ranap: boolean, noRawat: string, signal?: AbortSignal) {
  const query = new URLSearchParams({ no_rawat: noRawat })
  return request<DataRujukanInternal>(`${path(ranap)}?${query}`, {
    headers: { Authorization: `Bearer ${token}` },
    signal,
  })
}

export function cariReferensiRujukan(token: string, ranap: boolean, jenis: 'poli' | 'dokter', q: string) {
  return request<ReferensiRujukan[]>(`${path(ranap)}/referensi/${jenis}?${new URLSearchParams({ q })}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
}

export function simpanRujukanInternal(token: string, ranap: boolean, input: InputRujukan) {
  return request<{ pesan: string }>(path(ranap), {
    method: 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify(input),
  })
}

export function mutasiRujukanInternal(
  token: string,
  ranap: boolean,
  aksi: 'ubah' | 'hapus' | 'kirim',
  asal: RujukanInternal,
  baru?: InputRujukan,
) {
  return request<{ pesan: string; peringatan?: boolean }>(`${path(ranap)}${aksi === 'kirim' ? '/kirim' : ''}`, {
    method: aksi === 'ubah' ? 'PUT' : aksi === 'hapus' ? 'DELETE' : 'POST',
    headers: { Authorization: `Bearer ${token}` },
    body: JSON.stringify({ asal, baru }),
  })
}
