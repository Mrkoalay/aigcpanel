import {backend} from "./backend";

export type StorageBiz = "SoundPrompt" | "LiveAvatar" | "LiveKnowledge" | "LiveEvent" | "LiveTalk";

export type StorageRecord = {
    id?: number;

    biz: StorageBiz;

    sort?: number;
    title?: string;
    content?: any;

    runtime?: StorageRuntime;
};

export type StorageRuntime = {};

type StorageApiRecord = {
    id: number;
    biz: StorageBiz;
    sort: number;
    title: string;
    content: string;
    createdAt: number;
    updatedAt: number;
};

export const StorageService = {
    decodeRecord(record: StorageApiRecord | null): StorageRecord | null {
        if (!record) {
            return null;
        }
        return {
            ...record,
            content: JSON.parse(record.content ? record.content : "{}"),
        } as StorageRecord;
    },
    encodeRecord(record: StorageRecord): StorageApiRecord {
        return {
            id: Number(record.id || 0),
            biz: record.biz,
            sort: Number(record.sort || 0),
            title: record.title || "",
            content: JSON.stringify(record.content || {}),
            createdAt: 0,
            updatedAt: 0,
        };
    },
    async getByTitle(biz: StorageBiz, title: string): Promise<StorageRecord | null> {
        const records = await this.list(biz);
        return records.find(item => item.title === title) || null;
    },
    async get(id: number): Promise<StorageRecord | null> {
        const record = await backend.get<StorageApiRecord>(`/app/storages/${id}`);
        return this.decodeRecord(record);
    },
    async listByIds(ids: number[]): Promise<StorageRecord[]> {
        if (!ids || ids.length === 0) {
            return [];
        }
        const result: StorageRecord[] = [];
        for (const id of ids) {
            const record = await this.get(id);
            if (record) {
                result.push(record);
            }
        }
        return result;
    },
    async list(biz: StorageBiz): Promise<StorageRecord[]> {
        const records = await backend.get<StorageApiRecord[]>(`/app/storages?biz=${encodeURIComponent(biz)}`);
        return records.map(record => this.decodeRecord(record)!).sort((a, b) => (b.id || 0) - (a.id || 0));
    },
    async add(biz: StorageBiz, record: Partial<StorageRecord>) {
        record["biz"] = biz;
        const payload = this.encodeRecord(record as StorageRecord);
        const created = await backend.post<StorageApiRecord>("/app/storages", payload);
        return created.id;
    },
    async update(id: number, record: Partial<StorageRecord>) {
        const payload: any = {};
        if ("biz" in record) payload.biz = record.biz;
        if ("sort" in record) payload.sort = record.sort;
        if ("title" in record) payload.title = record.title;
        if ("content" in record) payload.content = JSON.stringify(record.content || {});
        await backend.patch(`/app/storages/${id}`, payload);
        return 1;
    },
    async addOrUpdate(biz: StorageBiz, id: number, record: Partial<StorageRecord>) {
        if (!id) {
            await this.add(biz, record);
        } else {
            await this.update(id, record);
        }
    },
    async delete(record: StorageRecord) {
        const filesForClean: string[] = [];
        if (record.content) {
            if (record.content.url) {
                filesForClean.push(record.content.url);
            }
        }
        for (const file of filesForClean) {
            await window.$mapi.file.deletes(file);
        }
        await backend.delete(`/app/storages/${record.id}`);
    },
    async clear(biz: StorageBiz) {
        await backend.delete(`/app/storages?biz=${encodeURIComponent(biz)}`);
    },
    async count(biz: StorageBiz, startTime: number = 0, endTime: number = 0) {
        const records = await this.list(biz);
        return records.filter(item => {
            const createdAt = (item as any).createdAt || 0;
            if (startTime > 0 && createdAt < startTime) {
                return false;
            }
            return !(endTime > 0 && createdAt > endTime);
        }).length;
    },
};
