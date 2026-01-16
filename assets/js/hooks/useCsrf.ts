export function getCsrfToken(): string {
  const match = document.cookie.match(/(?:^|; )_csrf=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : '';
}

export function useCsrf(): string {
  return getCsrfToken();
}
