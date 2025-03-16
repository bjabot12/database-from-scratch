# Auotmicity and Durability

Atomicity means that data is either updated or not, not in between. Durability means that data is guaranteed to exist after a certain point. They are not separate concerns, because we must achieved both.

The most basic way of achieving Automicity is creating a new temporary file, overwrite the old one if writing was successful and the deleting the old file. This will solve the problem if something wrong happens in the writing process, as we will always be able to restore the old version of the file, and therefore only losing the new data. Also, either all updates will happen or none will. So the user do not have to care if some of the updates occured during the write.

