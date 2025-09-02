import socket
import logging
import errno

from .server_protocol import ServerProtocol
from .utils import store_bets, Bet

class Server:
    IDX_AGENCY = 0
    IDX_FIRST_NAME = 1
    IDX_LAST_NAME = 2
    IDX_DOCUMENT = 3
    IDX_BIRTHDATE = 4
    IDX_NUMBER = 5
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._keep_running = True
        self.client_sock = None

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._keep_running:
            try:
                self.client_sock = self.__accept_new_connection()
                if self.client_sock:
                    server_protocol = ServerProtocol(self.client_sock, max_length=8192)
                    self.__handle_client_connection(server_protocol)
                    self.client_sock = None
            except OSError as e:
                if e.errno == errno.EBADF:
                    break
                raise

    def recv_bet(self, server_protocol):
        if campos := server_protocol.recv_bet():
            return campos
        logging.error("action: receive_message | result: fail | error: formato de mensaje incorrecto")
        return None

    def send_response_bet(self, server_protocol, nid, number):
        server_protocol.send_response(nid, number)

    def recv_batch(self, server_protocol):
        if batch := server_protocol.recv_batch():
            return batch
        return None

    def store_batch(self, batch):
        stored_count = 0
        for bet in batch:
            bet_obj = Bet(
                agency=bet[self.IDX_AGENCY],
                first_name=bet[self.IDX_FIRST_NAME],
                last_name=bet[self.IDX_LAST_NAME],
                document=bet[self.IDX_DOCUMENT],
                birthdate=bet[self.IDX_BIRTHDATE],
                number=bet[self.IDX_NUMBER]
            )
            try:
                store_bets([bet_obj])
                stored_count += 1
            except Exception as e:
                logging.error(f"action: store_batch | result: fail | error: {e}")
        return stored_count

    def send_response_batch(self, server_protocol, stored_count, amount_of_bets):
        if stored_count == amount_of_bets:
            server_protocol.send_response_batch("success")
        else:
            server_protocol.send_response_batch("fail")

    def __handle_client_connection(self, server_protocol):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            while self._keep_running:
                batch = self.recv_batch(server_protocol)
                if not batch:
                    # self.shutdown()
                    break
                amount_of_bets = len(batch)
                stored_count = self.store_batch(batch)
                if stored_count == amount_of_bets:
                    logging.info(f"action: apuesta_recibida | result: success | cantidad: {stored_count}")
                else:
                    logging.info(f"action: apuesta_recibida | result: fail | cantidad: {stored_count}")
                self.send_response_batch(server_protocol, stored_count, amount_of_bets)

        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            server_protocol.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def shutdown(self):
        self._keep_running = False
        if self.client_sock:
            self.client_sock.close()
        self._server_socket.close()
