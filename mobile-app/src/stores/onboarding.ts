import { create } from 'zustand'

type OnboardingMode = 'onboarding' | 'retake'
type OnboardingTransitionDirection = 'forward' | 'backward'

interface OnboardingSnapshot {
  markets?: string[]
  experience?: string
  style?: string
  painPoints?: string[]
  riskPreference?: string
}

interface OnboardingState {
  mode: OnboardingMode
  transitionDirection: OnboardingTransitionDirection
  markets: string[]
  experience: string
  style: string
  painPoints: string[]
  riskPreference: string
  setTransitionDirection: (value: OnboardingTransitionDirection) => void
  startRetake: (snapshot: OnboardingSnapshot) => void
  setMarkets: (values: string[]) => void
  setExperience: (value: string) => void
  setStyle: (value: string) => void
  setPainPoints: (values: string[]) => void
  setRiskPreference: (value: string) => void
  reset: () => void
}

const initialState: Pick<
  OnboardingState,
  'mode' | 'transitionDirection' | 'markets' | 'experience' | 'style' | 'painPoints' | 'riskPreference'
> = {
  mode: 'onboarding',
  transitionDirection: 'forward',
  markets: [],
  experience: '',
  style: '',
  painPoints: [],
  riskPreference: '',
}

export const useOnboardingStore = create<OnboardingState>((set) => ({
  ...initialState,
  setTransitionDirection: (value) => set({ transitionDirection: value }),
  startRetake: (snapshot) =>
    set({
      mode: 'retake',
      transitionDirection: 'forward',
      markets: snapshot.markets ?? [],
      experience: snapshot.experience ?? '',
      style: snapshot.style ?? '',
      painPoints: snapshot.painPoints ?? [],
      riskPreference: snapshot.riskPreference ?? '',
    }),
  setMarkets: (values) => set({ markets: values }),
  setExperience: (value) => set({ experience: value }),
  setStyle: (value) => set({ style: value }),
  setPainPoints: (values) => set({ painPoints: values }),
  setRiskPreference: (value) => set({ riskPreference: value }),
  reset: () => set(initialState),
}))
