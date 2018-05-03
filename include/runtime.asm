; Darwin

section .data
TRUE:   dq  -1                              ; True
FALSE:  dq  0                               ; False
EINVAL: db  'invalid input number', 0xa
ELEN:   equ $-EINVAL

section .bss
IOB: resb 1              ; I/O buffer used by READ and WRITE

section .text

; PERROR reports invalid input error and halts
PERROR:
    push dword ELEN      ; write the length of error msg
    push dword EINVAL    ; reference error msg to write
    push dword 1         ; file descriptor (stdout)
    sub esp, 4           ; darwin syscall need "extra space" on stack
    mov eax, 4           ; system call number (sys_write)
    int 0x80             ; call kernel

    call EXIT


; READ reads a byte from the standard input stream into the I/O buffer
READ:
    push eax             ; preserve eax, restore before procedure returns

    push dword 1
    push dword IOB
    push dword 0
    sub esp, 4
    mov eax, 3
    int 0x80

    add esp, 16

    pop eax
    ret


; WRITE writes the byte in al register to the standard output stream.
WRITE:
    push eax             ; preserve eax, restore before procedure returns
    mov byte [IOB], al   ; copy eax LO byte to i/o buffer

    push dword 1         ; write only one byte
    push dword IOB       ; reference buffer to write
    push dword 1         ; file descriptor (stdout)
    sub esp, 4           ; darwin syscall need "extra space" on stack
    mov eax, 4           ; system call number (sys_write)
    int 0x80             ; call kernel

    add esp, 16          ; clean stack (3 arguments * 4 + 4 bytes extra space)

    pop eax
    ret


; DRAIN reads the remaining bytes from the standard input stream
DRAIN:
    push ebx             ; preserve ebx, restore before procedure returns

    xor ebx, ebx

.next:
    call READ
    mov bl, byte [IOB]

    cmp bl, 0xa          ; ascii character '\n'
    je .done

    jmp .next

.done:

    pop ebx
    ret


; PRINTN writes the integer in the eax register to the standard output stream.
PRINTN:
    push eax            ; preserve eax, restore before procedure returns
    push ebp            ; preserve ebp, restore before procedure returns

    cmp eax, 0
    je .zero
    jg .positive

    ; integer is negative, write sign and negate
    push eax
    mov eax, 0x2d       ; ascii character '-'
    call WRITE
    pop eax
    neg eax

.positive:
    xor ebx, ebx        ; clear digits counter

.digits:
    cmp eax, 0
    je .convert

    xor edx, edx        ; clear reminder
    mov ecx, 10         ; divisor
    div ecx             ; divide, put quotient in eax and reminder in edx

    push edx            ; push digit onto stack
    inc ebx             ; increment digits counter

    jmp .digits    ; repeat

.convert:
    cmp ebx, 0
    je .done

    dec ebx
    pop eax
    add eax, 0x30       ; convert digit to ascii by adding '0' character
    call WRITE

    jmp .convert

.zero:
    mov eax, 0x30       ; copy ascii '0' character
    call WRITE

.done:
    pop ebp
    pop eax
    ret

; SCANN reads from the standard input stream a number into the eax register
SCANN:
    push ebx             ; preserve ebx, restore before procedure returns

    xor eax, eax
    xor ebx, ebx

.nextDigit:
    call READ
    mov bl, byte [IOB]

    ; EOF?
    cmp bl, 0xa
    je .done

    ; isDigit?
    cmp bl, '0'       ; < ascii '0' character
    jl .invalid
    cmp bl, '9'       ; > ascii '9' character
    jg .invalid

    imul eax, 10
    sub bl, '0'       ; convert to decimal representation by subtracting '0'
    add eax, ebx

    jmp .nextDigit

.invalid:
    call DRAIN
    call PERROR

.done:
    pop ebx
    ret


; NEWLINE writes a newline character ("\n") to the standard output stream.
NEWLINE:
    push eax             ; preserve eax, restore before procedure returns

    mov eax, 0x0A
    call WRITE

    pop eax
    ret

; EXIT returns control to the operating system
EXIT:
    push dword 0        ; exit code
    mov eax, 1          ; system call number (sys_exit)
    sub esp, 4          ; darwin syscall need "extra space" on stack
    int 0x80            ; call kernel

    ; no need to clean up stack after program exit
