class ServerProtocol:
    MAX_LENGTH = 1024
    def __init__(self, client_sock):
        self.client_sock = client_sock
        self.max_length = self.MAX_LENGTH

    def recv_all(self):
        buffer = b""
        while True:
            chunk = self.client_sock.recv(self.max_length-len(buffer))
            if not chunk:
                break 
            buffer += chunk
            if b"\n" in chunk:
                break
        msg = buffer.rstrip().decode('utf-8')
        return msg

    def recv_bet(self):
        msg = self.recv_all()
        campos = msg.split('|')
        if len(campos) == 6:
            return campos
        else:
            return None

    def send_response_bet(self, nid, number):
        response = f"{nid}|{number}\n"
        self.client_sock.sendall(response.encode('utf-8'))


    def close(self):
        self.client_sock.close()
