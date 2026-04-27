export const authState = $state({
  token: typeof window !== "undefined" ? localStorage.getItem("tether:token") || "" : "",
  isAuthed: false,
  setToken(t: string) {
    this.token = t;
    if (t) {
      localStorage.setItem("tether:token", t);
    } else {
      localStorage.removeItem("tether:token");
    }
  },
});
