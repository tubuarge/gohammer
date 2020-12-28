#!/usr/bin/env python3

# extend sys.path for imports:
if __name__ == '__main__' and __package__ is None:
	from os import sys, path
	sys.path.append(path.dirname(path.dirname(path.abspath(__file__))))

# standard library:
import time

# pypi:
import requests

# chainhammer:
import hammer
from hammer.config import Node1RPCaddress
from hammer.clienttype import curl_post

def call_port(RPCaddress=Node1RPCaddress):
	"""
	Just calls empty GET on RPCaddress, checks whether status_code==200
	"""
	headers=headers = {'Content-type': 'application/json'}
	try:
		response = requests.post(RPCaddress, headers=headers) # GET does not work on parity instantseal
	except requests.exceptions.ConnectionError:
		success, error = False, 'ConnectionError'
	else:
		success, error = True, None
		# print ("response.status_code =", response.status_code)
		if response.status_code!=200:
			success, error = False, "response.status_code = %d" % response.status_code
		# print ("response.text =", response.text)

	return success, error


def simple_RPC_call(RPCaddress=Node1RPCaddress, method="web3_clientVersion"):
	"""
	calls simplemost RPC call 'web3_clientVersion' and checks answer.
	returns (BOOL success, STRING/None error)
	"""
	try:
		answer = curl_post(method, RPCaddress=RPCaddress)
	except requests.exceptions.ConnectionError:
		success, error = False, "ConnectionError"
	except hammer.clienttype.MethodNotExistentError:
		success, error = False, "MethodNotExistentError"
	else:
		try:
			nodeName = answer.split("/")[0]
			#print (nodeName)
			success, error = True, None
		except Exception as e:
			success, error = False, "Exception: (%s) %s" % (type(e), e)

	return success, error


def loop_until_is_up(seconds_between_calls = 0.5, ifPrint=False, timeout=None):
	"""
	endless loop, until RPC API call answers something
	"""
	start = time.monotonic()
	while True:
		success, error = simple_RPC_call()
		# TODO: Also wait until chain is moving (for all but testRPC & instantseal)?
		# E.g. quorum-crux needs a long time between first answer and blocks moving!
		# or a simple sleep, depending on NODENAME ?
		if ifPrint:
			print (success, error)
		if success:
			print("Node is working.")
			break
		if timeout:
			if time.monotonic() - start > timeout:
				break
		time.sleep(seconds_between_calls)

	return success


if __name__ == '__main__':

	#print (call_port(RPCaddress=Node1RPCaddress))
	#print (simple_RPC_call())
	# loop_until_is_up(ifPrint=True)
	# loop_until_is_up(timeout=3)

	loop_until_is_up()

