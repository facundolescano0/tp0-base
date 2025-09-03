import socket
import logging


class ServerProtocol:
    HEADER = 0
    BATCH_FINISHED = "0"
    WINNERS_REQUEST = "1"
    ONE_BYTE = 1

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
            if b"\n\n" in chunk:
                break
        msg = buffer.decode('utf-8').rstrip('\n\n')
        logging.info(f"action: recv_all | message: ({msg})")
        return msg
    
    def recv_line(self):
        buffer = b""
        while True:
            chunk = self.client_sock.recv(ServerProtocol.ONE_BYTE)
            if not chunk:
                break
            buffer += chunk
            if chunk == b"\n":
                break
        return buffer.decode('utf-8').rstrip('\n')
        
    def recv_agency_id(self):
        return self.recv_line()

    def recv_winners_request(self):
        winner_request = self.recv_line()
        if winner_request == ServerProtocol.WINNERS_REQUEST:
            return winner_request
        logging.info(f"Received winners request: {winner_request} vs {ServerProtocol.WINNERS_REQUEST}")
        return None

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
        if data == ServerProtocol.BATCH_FINISHED:
            return ServerProtocol.BATCH_FINISHED
        if not data:
            logging.info("Received empty batch data")
            return None
        lines = data.splitlines()
        if not lines:
            logging.info("Not lines")
            return None
        
        header = lines[ServerProtocol.HEADER]

        try:
            bet_count = int(header.strip())
        except ValueError:
            logging.info("no puedo convertir a numero")
            return None

        if bet_count == ServerProtocol.BATCH_FINISHED:
                return ServerProtocol.BATCH_FINISHED
        
        if len(lines[1:]) < bet_count:
            logging.info("no hay suficientes lineas")
            return None

        batch = []
        for line in lines[1:1+bet_count]:
            campos = line.split('|')
            if len(campos) == 6:
                batch.append(campos)
            else:
                logging.info("no hay 6 campos")
                return None
        logging.info(f"este es el batch que te devuelve recv_Batch {batch}")
        return batch

    def send_response_batch(self, message):
        response = f"{message}\n"
        self.client_sock.sendall(response.encode('utf-8'))

    def send_winners(self, winners):
        if len(winners) == 0:
            logging.info("action: send_winners | result: no_winners")
            return None
        msg = ""
        for i, dni in enumerate(winners):
            if i > 0:
                msg += "|"
            msg += str(dni)
        msg += "\n"
        self.client_sock.sendall(msg.encode('utf-8'))

    def close(self):
        self.client_sock.close()
