const API_BASE = (import.meta.env.VITE_GO_API_BASE || "http://127.0.0.1:8080") + "/api/v1";

async function call<T>(path: string, init?: RequestInit): Promise<T> {
    const response = await fetch(`${API_BASE}${path}`, {
        headers: {
            "Content-Type": "application/json",
            ...(init?.headers || {}),
        },
        ...init,
    });
    if (!response.ok) {
        const text = await response.text();
        throw new Error(text || `request failed: ${response.status}`);
    }
    if (response.status === 204) {
        return undefined as T;
    }
    return await response.json() as T;
}

export const backend = {
    get: <T>(path: string) => call<T>(path),
    post: <T>(path: string, data?: any) =>
        call<T>(path, {
            method: "POST",
            body: JSON.stringify(data || {}),
        }),
    patch: <T>(path: string, data?: any) =>
        call<T>(path, {
            method: "PATCH",
            body: JSON.stringify(data || {}),
        }),
    delete: <T>(path: string) =>
        call<T>(path, {
            method: "DELETE",
        }),
};
