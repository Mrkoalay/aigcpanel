import {backend} from "./backend";

export type VideoTemplateRecord = {
    id?: number;
    name: string;
    video: string;
    info: any;
};

type VideoTemplateApiRecord = {
    id: number;
    name: string;
    video: string;
    info: string;
    createdAt: number;
    updatedAt: number;
};

export const VideoTemplateService = {
    decodeRecord(record: VideoTemplateApiRecord): VideoTemplateRecord | null {
        if (!record) {
            return null;
        }
        return {
            ...record,
            info: record.info ? JSON.parse(record.info) : {},
        } as VideoTemplateRecord;
    },
    encodeRecord(record: VideoTemplateRecord): any {
        return {
            ...record,
            info: JSON.stringify(record.info || {}),
        };
    },
    async get(id: number): Promise<VideoTemplateRecord | null> {
        const record = await backend.get<VideoTemplateApiRecord>(`/app/templates/${id}`);
        return this.decodeRecord(record);
    },
    async getByName(name: string): Promise<VideoTemplateRecord | null> {
        const record = await backend.get<VideoTemplateApiRecord>(`/app/templates/${encodeURIComponent(name)}`);
        return this.decodeRecord(record);
    },
    async list(): Promise<VideoTemplateRecord[]> {
        const records = await backend.get<VideoTemplateApiRecord[]>("/app/templates");
        return records.map(record => this.decodeRecord(record) as VideoTemplateRecord).sort((a, b) => (b.id || 0) - (a.id || 0));
    },
    async insert(record: VideoTemplateRecord) {
        const created = await backend.post<VideoTemplateApiRecord>("/app/templates", this.encodeRecord(record));
        return created.id;
    },
    async delete(record: VideoTemplateRecord) {
        if (record.video) {
            await window.$mapi.file.hubDelete(record.video);
        }
        await backend.delete(`/app/templates/${record.id}`);
    },
    async update(id: number, record: Partial<VideoTemplateRecord>) {
        const payload: any = {...record};
        if ("info" in payload) {
            payload.info = JSON.stringify(payload.info || {});
        }
        await backend.patch(`/app/templates/${id}`, payload);
        return 1;
    },
};
