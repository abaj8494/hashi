import numpy as np
import copy
import sys
from itertools import permutations
import os
from collections import deque
import time
import heapq

class Island:
    def __init__(self, row, col, is_island, value, bridges):
        self.row = row
        self.col = col
        self.is_island = is_island
        self.value = value
        self.bridges = bridges

    def __lt__(self, other):
        return self.value > other.value


def find_solution(board, viable_moves):
    moves = []
    bad_moves = []
    board_states = []
    for perm in permutations(viable_moves):
        board_copy = copy.deepcopy(board)
        move_queue = deque(perm)
        for move in move_queue:
            island1, island2 = move
            valid = perform_move(board_copy, island1.row, island1.col, island2.row, island2.col)
            if valid:
                moves.append([island1.row, island1.col, island2.row, island2.col])
            else:
                bad_moves.append((island1, island2))
            if is_winner(board_copy):
                return moves
        if bad_moves:
            for island1, island2 in bad_moves:
                valid = perform_move(board_copy, island1.row, island1.col, island2.row, island2.col)
                if valid:
                    moves.append([island1.row, island1.col, island2.row, island2.col])
        if is_winner(board_copy):
            return moves
        else:
            print_board(board_copy)
            board_states.append(board_copy)
            moves.clear()
            sys.exit(0)
    return None


def is_winner(board):
    rows, cols = board.shape
    islands = [board[i, j] for i in range(rows) for j in range(cols) if board[i, j].is_island]
    return all(isl.bridges == isl.value for isl in islands)


def create_board(file_path):
    # Read puzzle file: blank spots as '.', islands as digits or 'a'/'b'/'c' for 10/11/12
    with open(file_path, 'r') as f:
        n, m = map(int, f.readline().split(','))
        temp = []
        mapping = {'a': 10, 'b': 11, 'c': 12}
        for i in range(n):
            line = f.readline().strip()
            chars = list(line)
            row = []
            for j, ch in enumerate(chars):
                if ch == '.':
                    row.append(Island(i, j, False, '.', 0))
                else:
                    if ch.isdigit():
                        val = int(ch)
                    elif ch in mapping:
                        val = mapping[ch]
                    else:
                        raise ValueError(f"Invalid character in board: {ch}")
                    row.append(Island(i, j, True, val, 0))
            temp.append(row)
    # Expand grid for bridge placement
    expanded = []
    for i in range(n):
        expanded.append(temp[i])
        for j in range(m):
            if i < n - 1 and temp[i][j].is_island and temp[i+1][j].is_island:
                gap_row = [Island(i, j, False, '.', 0) for _ in range(len(temp[0]))]
                expanded.append(gap_row)
                break
    for i in range(len(expanded)):
        for j in range(len(expanded[i])):
            if j < len(expanded[i]) - 1 and expanded[i][j].is_island and expanded[i][j+1].is_island:
                for k in range(len(expanded)):
                    expanded[k].insert(j+1, Island(i, j, False, '.', 0))
                break
    board = np.array(expanded)
    update_positions(board)
    return board


def connect_islands(board, island1, island2, is_human=False):
    # Prevent exceeding allowed bridges
    for isl in (island1, island2):
        if isl.bridges == isl.value:
            if is_human:
                print("Maximum number of bridges reached")
            return
    # Horizontal
    if island1.row == island2.row:
        row = island1.row
        step = 1 if island1.col < island2.col else -1
        start = island1.col + step
        end = island2.col
        if board[row, start].bridges == 2:
            if is_human:
                print("Already two bridges")
            return
        for c in range(start, end, step):
            if board[row, c].bridges % 2 == 0:
                board[row, c].bridges = 1
                board[row, c].value = '-'
            else:
                board[row, c].bridges = 2
                board[row, c].value = '='
    # Vertical
    if island1.col == island2.col:
        col = island1.col
        step = 1 if island1.row < island2.row else -1
        start = island1.row + step
        end = island2.row
        if board[start, col].bridges == 2:
            if is_human:
                print("Already two bridges")
            return
        for r in range(start, end, step):
            if board[r, col].bridges % 2 == 0:
                board[r, col].bridges = 1
                board[r, col].value = '|'  
            else:
                board[r, col].bridges = 2
                board[r, col].value = '║'
    island1.bridges += 1
    island2.bridges += 1


def check_islands(board, island1, island2, is_human=False):
    # Ensure no other island between two
    if island1.row == island2.row:
        step = 1 if island1.col < island2.col else -1
        for c in range(island1.col + step, island2.col, step):
            if board[island1.row, c].is_island:
                if is_human:
                    print("Island in the way")
                return False
        return True
    if island1.col == island2.col:
        step = 1 if island1.row < island2.row else -1
        for r in range(island1.row + step, island2.row, step):
            if board[r, island1.col].is_island:
                if is_human:
                    print("Island in the way")
                return False
        return True
    if is_human:
        print("Cannot connect these islands")
    return False


def update_positions(board):
    rows, cols = board.shape
    for i in range(rows):
        for j in range(cols):
            board[i, j].row = i
            board[i, j].col = j


def validate_row(board, island1, island2, bridges):
    step = 1 if island1.col < island2.col else -1
    symbol = '-' if bridges == 1 else '.'
    for c in range(island1.col + step, island2.col, step):
        if board[island1.row, c].value != symbol:
            return False
    return True


def validate_col(board, island1, island2, bridges):
    step = 1 if island1.row < island2.row else -1
    symbol = '|' if bridges == 1 else '.'
    for r in range(island1.row + step, island2.row, step):
        if board[r, island1.col].value != symbol:
            return False
    return True


def perform_move(board, row1, col1, row2, col2):
    valid = False
    cell1 = board[row1, col1]
    cell2 = board[row2, col2]
    if cell1.is_island and cell2.is_island:
        if check_islands(board, cell1, cell2):
            # Next cell determines existing bridges
            if cell1.row == cell2.row:
                next_cell = board[cell1.row, cell1.col + (1 if cell1.col < cell2.col else -1)]
                valid = validate_row(board, cell1, cell2, next_cell.bridges)
            else:
                next_cell = board[cell1.row + (1 if cell1.row < cell2.row else -1), cell1.col]
                valid = validate_col(board, cell1, cell2, next_cell.bridges)
            if valid:
                connect_islands(board, cell1, cell2)
    return valid


def print_board(board):
    os.system('clear')
    rows, cols = board.shape
    for i in range(rows):
        for j in range(cols):
            print(f"  {board[i, j].value}  ", end="")
        print()
    time.sleep(1)


def move(board, row1, col1, row2, col2, is_human=False):
    if board[row1, col1].is_island and board[row2, col2].is_island:
        if check_islands(board, board[row1, col1], board[row2, col2], is_human):
            if perform_move(board, row1, col1, row2, col2):
                connect_islands(board, board[row1, col1], board[row2, col2], is_human)
    else:
        print("No island at given coordinates")


def automatic(board):
    possible_moves = []
    rows, cols = board.shape
    for i in range(rows):
        for j in range(cols):
            if board[i, j].is_island:
                # horizontal neighbors
                line = board[i, :]
                left, right = line[:j], line[j:]
                for k in range(len(left)-1, -1, -1):
                    if left[k].is_island:
                        possible_moves.append((board[i, j], left[k]))
                        break
                for neighbor in right:
                    if neighbor.is_island and neighbor.col != j:
                        possible_moves.append((board[i, j], neighbor))
                        break
                # vertical neighbors
                col_line = board[:, j]
                up, down = col_line[:i], col_line[i:]
                for k in range(len(up)-1, -1, -1):
                    if up[k].is_island:
                        possible_moves.append((board[i, j], up[k]))
                        break
                for neighbor in down:
                    if neighbor.is_island and neighbor.row != i:
                        possible_moves.append((board[i, j], neighbor))
                        break
    return find_solution(copy.deepcopy(board), possible_moves)


if __name__ == "__main__":
    board = create_board("islands.in")
    print_board(board)
    rows, cols = board.shape
    is_human = False
    print("Automatic (1)")
    print("Human (2)")
    while True:
        choice = input("How run? (1=auto, 2=human): ").strip()
        if choice == "1":
            break
        elif choice == "2":
            is_human = True
            break
    while True:
        try:
            if is_human:
                print_board(board)
                coords = list(map(int, input("Enter coords row1,col1,row2,col2: ").split(',')))
                if len(coords) != 4:
                    raise ValueError("Enter 4 comma-separated numbers")
                if any(c < 0 or c >= rows for c in (coords[0], coords[2])) or any(c < 0 or c >= cols for c in (coords[1], coords[3])):
                    raise ValueError("Invalid coordinates")
                move(board, *coords, True)
            else:
                solution = automatic(board)
                if solution is not None:
                    for step in solution:
                        print_board(board)
                        move(board, step[0], step[1], step[2], step[3])
                else:
                    print("No solution")
        except ValueError as e:
            print(e)
        if is_winner(board):
            print_board(board)
            print("You win!")
            break
