class ShelfAssignState {
  open = $state(false);
  bookIds = $state<string[]>([]);

  openFor(bookIds: string[]) {
    this.bookIds = bookIds;
    this.open = true;
  }

  close() {
    this.open = false;
    this.bookIds = [];
  }
}

export const shelfAssignState = new ShelfAssignState();
