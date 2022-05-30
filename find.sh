#!/bin/bash

a=1

until [ ! $a -lt 7 ]
do
   tree /tmp/$a
   a=`expr $a + 1`
done