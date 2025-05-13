import numpy as np
import copy
import sys
from itertools import permutations
from collections import deque
import heapq

class Island:
    def __init__(self, row, col, is_island, value, bridges=0):
        self.row = row
        self.col = col
        self.is_island = is_island
        self.value = value
        self.bridges = bridges

    def __lt__(self, other):
        return self.value > other.value


def read_board_stdin():
    lines = [line.rstrip('\n') for line in sys.stdin if line.strip()]
    n = len(lines)
    m = len(lines[0]) if n else 0
    temp = []
    mapping = {'a': 10, 'b': 11, 'c': 12}
    for i, line in enumerate(lines):
        row = []
        for j, ch in enumerate(line):
            if ch == '.':
                row.append(Island(i, j, False, '.'))
            else:
                if ch.isdigit():
                    val = int(ch)
                elif ch in mapping:
                    val = mapping[ch]
                else:
                    raise ValueError(f"Invalid character: {ch}")
                row.append(Island(i, j, True, val))
        temp.append(row)
    expanded = []
    for i in range(n):
        expanded.append(temp[i])
        for j in range(m):
            if i < n-1 and temp[i][j].is_island and temp[i+1][j].is_island:
                gap = [Island(i, j, False, '.') for _ in range(m)]
                expanded.append(gap)
                break
    for i in range(len(expanded)):
        for j in range(len(expanded[0]) - 1):
            if expanded[i][j].is_island and expanded[i][j+1].is_island:
                for row in expanded:
                    row.insert(j+1, Island(i, j, False, '.'))
                break
    board = np.array(expanded)
    update_positions(board)
    return board


def update_positions(board):
    rows, cols = board.shape
    for i in range(rows):
        for j in range(cols):
            board[i, j].row = i
            board[i, j].col = j


def check_clear(board, i1, i2, existing):
    # existing: number of bridges already present between these islands
    if i1.row == i2.row:
        step = 1 if i1.col < i2.col else -1
        sym = {1: '-', 2: '=', 3: 'E'}.get(existing, None)
        for c in range(i1.col+step, i2.col, step):
            if board[i1.row, c].value not in ('.', sym):
                return False
        return True
    if i1.col == i2.col:
        step = 1 if i1.row < i2.row else -1
        sym = {1: '|', 2: '║', 3: '#'}.get(existing, None)
        for r in range(i1.row+step, i2.row, step):
            if board[r, i1.col].value not in ('.', sym):
                return False
        return True
    return False


def connect(board, i1, i2):
    # draw one more bridge segment and symbol
    if i1.row == i2.row:
        step = 1 if i1.col < i2.col else -1
        for c in range(i1.col+step, i2.col, step):
            cell = board[i1.row, c]
            if cell.bridges < 3:
                cell.bridges += 1
                cell.value = {1: '-', 2: '=', 3: 'E'}[cell.bridges]
    else:
        step = 1 if i1.row < i2.row else -1
        for r in range(i1.row+step, i2.row, step):
            cell = board[r, i1.col]
            if cell.bridges < 3:
                cell.bridges += 1
                cell.value = {1: '|', 2: '║', 3: '#'}[cell.bridges]
    i1.bridges += 1
    i2.bridges += 1


def find_moves(board):
    moves = []
    rows, cols = board.shape
    for i in range(rows):
        for j in range(cols):
            cell = board[i, j]
            if cell.is_island:
                # horizontal neighbors
                for k in range(j-1, -1, -1):
                    if board[i, k].is_island:
                        moves.append((cell, board[i, k])); break
                for k in range(j+1, cols):
                    if board[i, k].is_island:
                        moves.append((cell, board[i, k])); break
                # vertical neighbors
                for k in range(i-1, -1, -1):
                    if board[k, j].is_island:
                        moves.append((cell, board[k, j])); break
                for k in range(i+1, rows):
                    if board[k, j].is_island:
                        moves.append((cell, board[k, j])); break
    return moves


def solve(board):
    candidates = find_moves(board)
    for perm in permutations(candidates):
        b = copy.deepcopy(board)
        ok = True
        for i1, i2 in perm:
            # determine existing bridges count for this pair
            if i1.row == i2.row:
                existing = b[i1.row, i1.col + (1 if i1.col < i2.col else -1)].bridges
            else:
                existing = b[i1.row + (1 if i1.row < i2.row else -1), i1.col].bridges
            if not check_clear(b, i1, i2, existing):
                ok = False
                break
            connect(b, i1, i2)
        if ok:
            return b
    return None


def print_solution(board):
    for row in board:
        print(''.join(str(cell.value) for cell in row))


if __name__ == '__main__':
    board = read_board_stdin()
    solution = solve(board)
    if solution is None:
        print('No solution')
    else:
        print_solution(solution)

