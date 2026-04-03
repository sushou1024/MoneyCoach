import { create } from 'zustand'

type UploadDraft = {
  images: { imageId: string; uri: string }[]
}

interface UploadDraftState {
  drafts: Record<string, UploadDraft>
  setDraft: (batchId: string, images: { imageId: string; uri: string }[]) => void
  clearDraft: (batchId: string) => void
}

export const useUploadDraftStore = create<UploadDraftState>((set) => ({
  drafts: {},
  setDraft: (batchId, images) =>
    set((state) => ({
      drafts: {
        ...state.drafts,
        [batchId]: { images },
      },
    })),
  clearDraft: (batchId) =>
    set((state) => {
      const next = { ...state.drafts }
      delete next[batchId]
      return { drafts: next }
    }),
}))
