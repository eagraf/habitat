type ErrorState = {
    state: "error",
    message: string,
}

type LoadingState = {
    state: "loading",
}

type SuccessState<T> = {
    state: "success",
    data: T,
}

export type AsyncState<T> = LoadingState | ErrorState | SuccessState<T> 
