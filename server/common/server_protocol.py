import socket


class ServerProtocol:
    HEADER = 0
    def __init__(self, client_sock, max_length):
        self.client_sock = client_sock
        self.max_length = max_length

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

    def recv_batch(self):
        data = self.recv_all()
        if not data:
            return None
        lines = data.splitlines()
        if not lines:
            return None
        header = lines[ServerProtocol.HEADER]

        try:
            bet_count = int(header.strip())
        except ValueError:
            return None
        
        if len(lines[1:]) < bet_count:
            return None

        batch = []
        for line in lines[1:1+bet_count]:
            campos = line.split('|')
            if len(campos) == 6:
                batch.append(campos)
            else:
                return None
            
        return batch

    def send_response_batch(self, message):
        response = f"{message}\n"
        self.client_sock.sendall(response.encode('utf-8'))
        

    def close(self):
        self.client_sock.close()
