type ErrorState = {
    state: "error",
    message: string,
}

type InitState = {
    state: "init",
}

type LoadingState = {
    state: "loading",
}

type SuccessState<T> = {
    state: "success",
    data: T,
}

export type AsyncState<T> = InitState | LoadingState | ErrorState | SuccessState<T> 
