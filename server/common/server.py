import socket
import logging
import errno

from .server_protocol import ServerProtocol
from .utils import store_bets, Bet, load_bets, has_won

class Server:
    IDX_AGENCY = 0
    IDX_FIRST_NAME = 1
    IDX_LAST_NAME = 2
    IDX_DOCUMENT = 3
    IDX_BIRTHDATE = 4
    IDX_NUMBER = 5
    FINISHED_BATCHES = "0"
    def __init__(self, port, listen_backlog, clients_amount):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._keep_running = True
        self.client_sock = None
        self.agencies_sent_all = 0
        self.clients_amount = clients_amount

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        while self._keep_running:
            self.client_sock = self.__accept_new_connection()
            if self.client_sock:
                server_protocol = ServerProtocol(self.client_sock, max_length=8192)
                self.__handle_client_connection(server_protocol)
                self.client_sock = None
        self.shutdown()

    def recv_bet(self, server_protocol):
        if campos := server_protocol.recv_bet():
            return campos
        logging.error("action: receive_message | result: fail | error: formato de mensaje incorrecto")
        return None

    def send_response_bet(self, server_protocol, nid, number):
        server_protocol.send_response(nid, number)

    def recv_batch(self, server_protocol):
        return server_protocol.recv_batch()

    def store_batch(self, batch):
        stored_count = 0
        for bet in batch:
            bet_obj = Bet(
                agency=int(bet[self.IDX_AGENCY]),
                first_name=bet[self.IDX_FIRST_NAME],
                last_name=bet[self.IDX_LAST_NAME],
                document=bet[self.IDX_DOCUMENT],
                birthdate=bet[self.IDX_BIRTHDATE],
                number=int(bet[self.IDX_NUMBER])
            )
            try:
                store_bets([bet_obj])
                stored_count += 1
            except Exception as e:
                logging.error(f"action: store_batch | result: fail | error: {e}")
        return stored_count

    def send_response_batch(self, server_protocol, stored_count, amount_of_bets):
        if stored_count == amount_of_bets:
            server_protocol.send_response_batch(ServerProtocol.BATCH_OK)
        else:
            server_protocol.send_response_batch(ServerProtocol.BATCH_FAIL)

    def recv_winners_request(self, server_protocol):
        return server_protocol.recv_winners_request()


    def send_winners(self, agency_id, server_protocol):
        bets = load_bets()
        winners = [bet.document for bet in bets if bet.agency == agency_id and has_won(bet)]
        server_protocol.send_winners(winners)

    def __handle_client_connection(self, server_protocol):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            
            while self._keep_running:

                opcode = server_protocol.recv_opcode()

                if opcode == ServerProtocol.BATCH:
                    batch = self.recv_batch(server_protocol)
                    # if not batch:
                        # logging.info(f"action: receive_message | result: fail | error: batch nill")
                        # # self.shutdown()
                        # break
                    if batch == ServerProtocol.BATCH_FINISHED:
                        self.agencies_sent_all += 1
                        if self.agencies_sent_all == self.clients_amount:
                            logging.info("action: sorteo | result: success")
                        break
                    amount_of_bets = len(batch)
                    stored_count = self.store_batch(batch)
                    if stored_count == amount_of_bets:
                         logging.info(f"action: apuesta_recibida | result: success | cantidad: {stored_count}")
                    else:
                         logging.info(f"action: apuesta_recibida | result: fail | cantidad: {stored_count}")
                    self.send_response_batch(server_protocol, stored_count, amount_of_bets)
                    
                elif opcode == ServerProtocol.WINNERS_REQUEST:
                    agency_id = server_protocol.recv_agency_id()
                    
                    if self.agencies_sent_all == self.clients_amount:
                        winners_sent = self.send_winners(agency_id, server_protocol)
                    else:
                        server_protocol.send_not_ready()
                    break
                else:
                    logging.error(f"action: receive_message | result: fail | error: unexpected opcode: {opcode}")
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

        try:
            # Connection arrived
            logging.info('action: accept_connections | result: in_progress')
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except OSError as e:
            logging.error(f'action: accept_connections | result: fail | error: {e}')

    def shutdown(self):
        self._keep_running = False
        if self.client_sock:
            self.client_sock.close()
        self._server_socket.close()
